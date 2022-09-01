package images

import (
	"context"
	"fmt"
	"io/ioutil"
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

	// argo image provider ONLY understands Argo helm based applications
	// obviously we can expand upon this in future...

	for _, p := range a.paths {

		data, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, errors.Wrap(err, "error loading YAML file")
		}

		var app ArgoApplication
		if err = yaml.Unmarshal(data, &app); err != nil {
			return nil, errors.Wrap(err, "error unmarshalling YAML")
		}

		if app.Spec.Source.Helm == nil {
			return nil, errors.New("error finding Helm definition defined")
		}

		directory, err := ioutil.TempDir(".", "prefix")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(directory)

		if _, err = exec.GetCommandOutput(directory, "git", "clone", app.Spec.Source.RepoURL, "foo"); err != nil {
			return nil, errors.Wrap(err, "error running git clone")
		}

		directory = path.Join(directory, "foo")

		if _, err = exec.GetCommandOutput(directory, "git", "checkout", app.Spec.Source.TargetRevision); err != nil {
			return nil, errors.Wrap(err, "error running git checkout")
		}

		args := []string{
			"template",
		}

		for _, p := range app.Spec.Source.Helm.Parameters {
			args = append(args, "--set", fmt.Sprintf("%s=%s", p.Name, p.Value))
		}

		args = append(args, "foo")
		args = append(args, app.Spec.Source.Path)

		output, err := exec.GetCommandOutput(directory, "helm", args...)
		if err != nil {
			return nil, errors.Wrap(err, "error running helm template")
		}

		if err = splitYAMLAndRunXPath(output, "$..spec.containers[*].image", all); err != nil {
			return nil, errors.Wrap(err, "error extracting images")
		}
	}
	return all, nil

}
