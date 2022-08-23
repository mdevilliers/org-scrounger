package images

import (
	"context"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/util"
	"github.com/pkg/errors"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

type kustomize struct {
	paths []string
}

func NewKustomize(paths ...string) *kustomize {
	return &kustomize{
		paths: paths,
	}
}

func (k *kustomize) Images(ctx context.Context) (util.Set[string], error) {
	all := util.NewSet[string]()

	for _, root := range k.paths {

		elements, err := runKustomizeAndSelect(root, "$..spec.containers[*].image")
		if err != nil {
			return all, err
		}
		// add to all keeping count
		for _, element := range elements {
			i := strings.TrimSpace(element.Value)
			if i != "" {
				all.Add(i)
			}
		}
	}
	return all, nil
}

func runKustomizeAndSelect(directory, xpath string) ([]*yaml.Node, error) {
	ret := []*yaml.Node{}

	// run kustomize in root - get back big ball of yaml
	output, err := exec.GetCommandOutput(directory, "kustomize", "build")
	if err != nil {
		return nil, errors.Wrap(err, "error running kustomize")
	}
	// split out to the individual documents
	yamls := strings.Split(output, "\n---\n")

	for _, yamlstr := range yamls {
		// extract all the .image values
		var n yaml.Node

		if err := yaml.Unmarshal([]byte(yamlstr), &n); err != nil {
			return nil, errors.Wrap(err, "error unmarshalling kustomize output")
		}

		path, err := yamlpath.NewPath(xpath)
		if err != nil {
			return nil, errors.Wrap(err, "error creating yaml path")
		}
		elements, err := path.Find(&n)
		if err != nil {
			return nil, errors.Wrap(err, "error finding image nodes")
		}
		ret = append(ret, elements...)
	}
	return ret, nil
}
