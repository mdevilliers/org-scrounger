package images

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type argoProvider struct {
	deleteCacheOnExit bool
	paths             []string
}

func NewArgo(deleteCacheOnExit bool, paths ...string) *argoProvider {
	return &argoProvider{
		deleteCacheOnExit: deleteCacheOnExit,
		paths:             paths,
	}
}

type ArgoApplication struct {
	Kind string
	Spec struct {
		Destination struct {
			Namespace string
		}
		Source struct {
			RepoURL        string `yaml:"repoURL"`
			Path           string
			TargetRevision string `yaml:"targetRevision"`
			Helm           *struct {
				ReleaseName string
				Parameters  []struct {
					Name  string
					Value string
				}
				Values string // :shrug
			}
		}
	}
}

func (a *argoProvider) Images(ctx context.Context) ([]mapping.Image, error) {
	all := []mapping.Image{}

	// use the temp dir as the root of any checkout
	directory := os.TempDir()
	directory = path.Join(directory, "scrng")

	if a.deleteCacheOnExit {
		defer func() {
			os.RemoveAll(directory)
		}()
	}

	for _, p := range a.paths {

		data, err := os.ReadFile(p)
		if errors.Is(err, os.ErrNotExist) {
			log.Info().Msgf("file does not exist: %s", p)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("error loading YAML file: %w", err)
		}

		var app ArgoApplication
		if err = yaml.Unmarshal(data, &app); err != nil {
			return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
		}

		root, err := cachedGithubCheckout(directory, app.Spec.Source.RepoURL, app.Spec.Source.TargetRevision)
		if err != nil {
			return nil, fmt.Errorf("error checking out %s@%s: %w", app.Spec.Source.RepoURL, app.Spec.Source.TargetRevision, err)
		}
		var content string

		if app.Spec.Source.Helm != nil {
			content, err = runHelm(root, app)
			if err != nil {
				return nil, fmt.Errorf("error running helm template: %s, output: %s : %w", root, content, err)
			}
		} else {

			// assume there is a kustomise file available
			p := path.Join(root, app.Spec.Source.Path)
			content, err = runKustomize(p)
			if err != nil {
				return nil, fmt.Errorf("error running kustomize: %w", err)
			}
		}

		images, err := resolveImages(content, app.Spec.Destination.Namespace)
		if err != nil {
			return nil, fmt.Errorf("error extracting images: %w", err)
		}

		all = append(all, images...)

	}
	return all, nil
}

// runs Helm in the root directory returning the inflated output or an error
func runHelm(root string, app ArgoApplication) (string, error) {

	args := []string{
		"template",
	}

	for _, p := range app.Spec.Source.Helm.Parameters {

		v := p.Value

		// in Helm, comma separated strings need to be escaped and wrapped in "" :shrug
		if strings.Index(p.Value, ",") > 0 {
			v = strings.ReplaceAll(p.Value, ",", "\\,")
			v = fmt.Sprintf("\"%s\"", v)
		}

		args = append(args, "--set", fmt.Sprintf("%s=%s", p.Name, v))
	}
	if app.Spec.Source.Helm.Values != "" {

		// create temp file with contents
		f, err := os.CreateTemp("", "values-file")
		if err != nil {
			return "", fmt.Errorf("error creating tempfile: %w", err)
		}
		defer os.Remove(f.Name())

		if _, err = f.WriteString(app.Spec.Source.Helm.Values); err != nil {
			return "", fmt.Errorf("error writing to  tempfile: %w", err)
		}
		args = append(args, "-f", f.Name())
	}

	args = append(args, "foo")
	args = append(args, app.Spec.Source.Path)

	return exec.GetCommandOutput(root, "helm", args...)
}

// clones the githubURL and checkouts the 'tagOrHead' returing the directory path or an error
func cachedGithubCheckout(directory string, githubURL, tagOrHead string) (string, error) {

	githubFolderName := strings.ReplaceAll(githubURL, "/", "_")
	folder := fmt.Sprintf("./%s-%s", githubFolderName, tagOrHead)
	p := path.Join(directory, folder)

	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating checkout directory: %w", err)
	}

	if stat, err := os.Stat(p); err == nil && stat.IsDir() {
		return p, nil
	}

	if _, err := exec.GetCommandOutput(directory, "git", "clone", githubURL, folder); err != nil {
		return "", fmt.Errorf("error running git clone: %w", err)
	}

	if _, err := exec.GetCommandOutput(p, "git", "checkout", tagOrHead); err != nil {
		return "", fmt.Errorf("error running git checkout: %w", err)
	}

	return p, nil
}
