package cmds

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/util"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func ImagesCmd() *cli.Command {
	return &cli.Command{
		Name: "images",
		Subcommands: []*cli.Command{
			&cli.Command{
				Name: "kustomize",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "root",
						Aliases: []string{"r"},
						Usage:   "path to root of kustomize config",
					},
					&cli.BoolFlag{
						Name:  "omit-usage-count",
						Value: false,
						Usage: "omit usage count",
					},
				},
				Action: func(c *cli.Context) error {

					roots := c.Value("root").(cli.StringSlice)
					omitUsageCount := c.Value("omit-usage-count").(bool)
					all := util.NewSet[string]()

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
								image := element.Value
								all.Add(image)
							}
						}
					}
					// sort alphabetially by key
					keys := make([]string, 0, len(all))
					for k := range all {
						keys = append(keys, k)
					}
					sort.Strings(keys)

					for _, key := range keys {
						if omitUsageCount {
							fmt.Println(key)
						} else {
							fmt.Println(key, all[key])
						}
					}

					return nil
				},
			},
		},
	}
}
