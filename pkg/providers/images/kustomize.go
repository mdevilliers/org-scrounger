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
		if err := runKustomizeAndSelect(root, "$..spec.containers[*].image", all); err != nil {
			return nil, err
		}
	}
	return all, nil
}

func runKustomizeAndSelect(directory, xpath string, set util.Set[string]) error {
	// run kustomize in root - get back big ball of yaml
	output, err := exec.GetCommandOutput(directory, "kustomize", "build")
	if err != nil {
		return errors.Wrap(err, "error running kustomize")
	}
	return splitYAMLAndRunXPath(output, xpath, set)
}

func splitYAMLAndRunXPath(output string, xpath string, set util.Set[string]) error {
	// split out to the individual documents
	yamls := strings.Split(output, "\n---\n")

	for _, yamlstr := range yamls {
		// extract all the .image values
		var n yaml.Node

		if err := yaml.Unmarshal([]byte(yamlstr), &n); err != nil {
			return errors.Wrap(err, "error unmarshalling kustomize output")
		}

		path, err := yamlpath.NewPath(xpath)
		if err != nil {
			return errors.Wrap(err, "error creating yaml path")
		}
		elements, err := path.Find(&n)
		if err != nil {
			return errors.Wrap(err, "error finding image nodes")
		}
		for _, element := range elements {
			i := strings.TrimSpace(element.Value)
			if i != "" {
				set.Add(i)
			}
		}
	}
	return nil
}
