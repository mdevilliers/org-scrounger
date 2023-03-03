package images

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
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
			return nil, fmt.Errorf("error running kustomize: %w", err)
		}
		images, err := resolveImages(content, "unknown")
		if err != nil {
			return nil, fmt.Errorf("error extracting images: %w", err)
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
func resolveImages(probablyYaml, namespace string) ([]mapping.Image, error) { //nolint:funlen

	all := map[string]mapping.Image{}

	// split out to the individual documents
	yamls := strings.Split(probablyYaml, "\n---\n")

	for _, yamlstr := range yamls {
		// extract all the .image values
		var n yaml.Node

		if err := yaml.Unmarshal([]byte(yamlstr), &n); err != nil {
			return nil, fmt.Errorf("error unmarshalling yaml: %w", err)
		}

		namespaceElement, err := compileAndExecuteXpath("$..metadata.namespace", &n)
		if err != nil {
			return nil, fmt.Errorf("error running namespace xpath: %w", err)
		}
		if len(namespaceElement) == 1 {
			namespace = namespaceElement[0].Value
		}

		specElements, err := compileAndExecuteXpath("$..spec", &n)
		if err != nil {
			return nil, fmt.Errorf("error running spec xpath: %w", err)
		}
		for _, spec := range specElements {
			imageElements, err := compileAndExecuteXpath("$..image", spec)
			if err != nil {
				return nil, fmt.Errorf("error running image xpath: %w", err)
			}
			if len(imageElements) == 0 {
				continue
			}
			// REVIEW : review this logic
			replicas, err := parseReplicaCount(spec)
			if err != nil {
				return nil, fmt.Errorf("error running replicas xpath: %w", err)
			}
			for _, element := range imageElements {
				image, version := splitImageAndVersion(strings.TrimSpace(element.Value))
				i := mapping.Image{
					Name:    image,
					Version: version,
					Count:   replicas,
					Destination: &mapping.Destination{
						Namespace: namespace,
					},
				}
				key := fmt.Sprintf("%s_%s", i.Name, i.Version)
				v, exists := all[key]
				if !exists {
					all[key] = i
				} else {
					v.Count++
					all[key] = v
				}
			}
		}
	}

	ret := []mapping.Image{}
	for _, v := range all {
		ret = append(ret, v)
	}

	return ret, nil
}

func parseReplicaCount(n *yaml.Node) (int, error) {
	replicasElement, err := compileAndExecuteXpath(".replicas", n)
	if err != nil {
		return 0, fmt.Errorf("error running replicas xpath: %w", err)
	}
	if len(replicasElement) == 1 {
		return strconv.Atoi(replicasElement[0].Value)
	}
	// somethingelse but don't blow up
	return 1, nil
}

func compileAndExecuteXpath(xpath string, document *yaml.Node) ([]*yaml.Node, error) {
	path, err := yamlpath.NewPath(xpath)
	if err != nil {
		return nil, fmt.Errorf("error creating yaml path: %w", err)
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
	// deal with shas with versions
	if len(bits) == 3 { //nolint: gomnd
		version = fmt.Sprintf("%s:%s", bits[1], bits[2])
	}
	return imageName, version
}
