package images

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type argoProvider struct {
	paths []string
}

func NewArgo(paths ...string) *argoProvider {
	return &argoProvider{
		paths: paths,
	}
}

type ArgoApplication struct {
	Kind string
	Spec struct {
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
			}
		}
	}
}

func (a *argoProvider) Images(ctx context.Context) (util.Set[string], error) {
	all := util.NewSet[string]()

	// use the temp dir as the root of any checkout
	directory := os.TempDir()

	for _, p := range a.paths {

		data, err := os.ReadFile(p)
		if err != nil {
			return nil, errors.Wrap(err, "error loading YAML file")
		}

		var app ArgoApplication
		if err = yaml.Unmarshal(data, &app); err != nil {
			return nil, errors.Wrap(err, "error unmarshalling YAML")
		}

		root, err := cachedGithubCheckout(directory, app.Spec.Source.RepoURL, app.Spec.Source.TargetRevision)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking out %s@%s", app.Spec.Source.RepoURL, app.Spec.Source.TargetRevision)
		}

		// if it is Help we only support inlined variables
		if app.Spec.Source.Helm != nil {

			args := []string{
				"template",
			}

			for _, p := range app.Spec.Source.Helm.Parameters {
				args = append(args, "--set", fmt.Sprintf("%s=%s", p.Name, p.Value))
			}

			args = append(args, "foo")
			args = append(args, app.Spec.Source.Path)

			content, err := exec.GetCommandOutput(root, "helm", args...)
			if err != nil {
				return nil, errors.Wrap(err, "error running helm template")
			}

			if err = splitYAMLAndRunXPath(content, "$..spec.containers[*].image", all); err != nil {
				return nil, errors.Wrap(err, "error extracting images")
			}

		} else {

			// assume there is a kustomise file available
			if err := runKustomizeAndSelect(root, "$..spec.containers[*].image", all); err != nil {
				return nil, errors.Wrap(err, "error running kustomize")
			}

		}
	}
	return all, nil

}

// Clones the githubURL and checkouts the 'tagOrHead' returing the directory path or an error
func cachedGithubCheckout(directory string, githubURL, tagOrHead string) (string, error) {

	folder := fmt.Sprintf("./%s-%s", githubURL, tagOrHead)
	p := path.Join(directory, folder)

	if stat, err := os.Stat(p); err == nil && stat.IsDir() {
		return p, nil
	}

	if _, err := exec.GetCommandOutput(directory, "git", "clone", githubURL, folder); err != nil {
		return "", errors.Wrap(err, "error running git clone")
	}

	if _, err := exec.GetCommandOutput(p, "git", "checkout", tagOrHead); err != nil {
		return "", errors.Wrap(err, "error running git checkout")
	}

	return p, nil
}
