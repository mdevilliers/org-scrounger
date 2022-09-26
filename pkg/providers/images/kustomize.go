package images

import (
	"context"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
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

func (k *kustomize) Images(ctx context.Context) ([]mapping.Image, error) {
	all := []mapping.Image{}

	for _, path := range k.paths {
		content, err := runKustomize(path)
		if err != nil {
			return nil, errors.Wrap(err, "error running kustomize")
		}
		images, err := resolveImages(content, "unknown")
		if err != nil {
			return nil, errors.Wrap(err, "error extracting images")
		}

		all = append(all, images...)

	}
	return all, nil
}

// runKustomize shells out to a directory and returns a
// big ball of yaml or an error
func runKustomize(directory string) (string, error) {
	return exec.GetCommandOutput(directory, "kustomize", "build")
}

// resolveImages produces a slice of Images or an error
func resolveImages(probablyYaml, namespace string) ([]mapping.Image, error) {

	all := []mapping.Image{}

	// split out to the individual documents
	yamls := strings.Split(probablyYaml, "\n---\n")

	for _, yamlstr := range yamls {
		// extract all the .image values
		var n yaml.Node

		if err := yaml.Unmarshal([]byte(yamlstr), &n); err != nil {
			return nil, errors.Wrap(err, "error unmarshalling yaml")
		}

		namespaceElement, err := compileAndExecuteXpath("$..metadata.namespace", &n)
		if err != nil {
			return nil, errors.Wrap(err, "error running namespace xpath")
		}
		if len(namespaceElement) == 1 {
			namespace = namespaceElement[0].Value
		}

		imageElements, err := compileAndExecuteXpath("$..spec.containers[*].image", &n)
		if err != nil {
			return nil, errors.Wrap(err, "error running image xpath")
		}
		for _, element := range imageElements {

			image, version := splitImageAndVersion(strings.TrimSpace(element.Value))

			all = append(all, mapping.Image{
				Name:    image,
				Version: version,
				Count:   1, // TODO: parse out the replicaCount value
				Destination: &mapping.Destination{
					Namespace: namespace,
				},
			})
		}

	}
	return all, nil
}

func compileAndExecuteXpath(xpath string, document *yaml.Node) ([]*yaml.Node, error) {
	path, err := yamlpath.NewPath(xpath)
	if err != nil {
		return nil, errors.Wrap(err, "error creating yaml path")
	}
	return path.Find(document)
}

func splitImageAndVersion(name string) (string, string) {

	bits := strings.Split(name, ":")

	imageName := bits[0]
	version := "unknown"
	if len(bits) == 2 { //nolint: gomnd
		version = bits[1]
	}
	return imageName, version
}
