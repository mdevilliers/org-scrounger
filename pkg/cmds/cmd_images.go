package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
	"github.com/mdevilliers/org-scrounger/pkg/providers/images"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type (
	Image struct {
		Name    string             `json:"name"`
		Version string             `json:"version"`
		Count   int                `json:"count"`
		Repo    *gh.RepositorySlim `json:"repo,omitempty"`
	}
)

func imagesCmd() *cli.Command { // nolint:funlen
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

					ghClient := gh.NewClientFromEnv(c.Context)

					kustomize := images.NewKustomize(roots.Value()...)
					all, err := kustomize.Images()

					if err != nil {
						return err // already wrapped
					}

					var (
						mapper *mapping.Mapper
					)

					if mappingFile != "" {
						mapper, err = mapping.LoadFromFile(mappingFile, ghClient)
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
