package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
	"github.com/mdevilliers/org-scrounger/pkg/util"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func imagesCmd() *cli.Command { // nolint: funlen
	return &cli.Command{
		Name: "images",
		Subcommands: []*cli.Command{
			{
				Name: "kustomize",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "root",
						Aliases: []string{"r"},
						Usage:   "path to root of kustomize config",
					},
					&cli.StringFlag{
						Name:  "mapping",
						Usage: "path to a mapping file",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: JSONOutputStr,
						Usage: fmt.Sprintf("specify output format [%s]. Default is '%s'.", JSONOutputStr, JSONOutputStr),
					},
				},
				Action: func(c *cli.Context) error {

					roots := c.Value("root").(cli.StringSlice)
					mappingFile := c.Value("mapping").(string)
					output := c.Value("output").(string)
					all := util.NewSet[string]()

					ghClient := gh.NewClientFromEnv(c.Context)

					for _, root := range roots.Value() {

						// run kustomize in root - get back big ball of yaml
						output, err := exec.GetCommandOutput(root, "kustomize", "build")
						if err != nil {
							return errors.Wrap(err, "error running kustomize")
						}
						// split out to the individual documents
						yamls := strings.Split(output, "\n---\n")

						for _, yamlstr := range yamls {
							// extract all the .image values
							var n yaml.Node

							if err := yaml.Unmarshal([]byte(yamlstr), &n); err != nil {
								return errors.Wrap(err, "error unmarshalling kustomize output")
							}

							path, err := yamlpath.NewPath("$..spec.containers[*].image")
							if err != nil {
								return errors.Wrap(err, "error creating yaml path")
							}

							elements, err := path.Find(&n)
							if err != nil {
								return errors.Wrap(err, "error finding image nodes")
							}

							// add to all keeping count
							for _, element := range elements {
								i := strings.TrimSpace(element.Value)
								if i != "" {
									all.Add(i)
								}
							}
						}
					}

					type (
						Image struct {
							Name    string             `json:"name"`
							Version string             `json:"version"`
							Count   int                `json:"count"`
							Repo    *gh.RepositorySlim `json:"repo,omitempty"`
						}
					)

					var (
						mapper *mapping.Mapper
					)

					if mappingFile != "" {

						file, err := os.Open(mappingFile)
						if err != nil {
							return errors.Wrapf(err, "error opening mapping file : %s", mappingFile)
						}
						rules, err := parser.UnMarshal(mappingFile, file)
						if err != nil {
							return errors.Wrap(err, "error reading mapping file")
						}

						mapper, err = mapping.New(rules, ghClient)
						if err != nil {
							return errors.Wrap(err, "error creating mapper")
						}
					}

					images := []Image{}

					for _, key := range all.OrderedKeys() {
						bits := strings.Split(key, ":")

						imageName := bits[0]
						version := "unknown"
						if len(bits) == 2 { // nolint: gomnd
							version = bits[1]
						}

						image := Image{
							Name:    imageName,
							Version: version,
							Count:   all[key],
						}

						if mapper != nil {
							found, repo, err := mapper.RepositoryFromImage(bits[0])
							if err != nil {
								return errors.Wrapf(err, "error mapping image '%s' to repo", bits[0])
							}
							if found {
								image.Repo = &repo
							}
						}
						images = append(images, image)
					}
					switch output {
					case JSONOutputStr:
						b, err := json.Marshal(images)
						if err != nil {
							return errors.Wrap(err, "error marshalling to json")
						}
						os.Stdout.Write(b)
					default:
						return errors.New("unknown output")
					}

					return nil
				},
			},
		},
	}
}
