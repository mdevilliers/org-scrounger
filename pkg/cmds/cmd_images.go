package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
	"github.com/mdevilliers/org-scrounger/pkg/providers/images"
	"github.com/mdevilliers/org-scrounger/pkg/sonarcloud"
	"github.com/mdevilliers/org-scrounger/pkg/util"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type imageProvider interface {
	Images() (util.Set[string], error)
}

func imagesCmd() *cli.Command { //nolint:funlen
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
					kustomize := images.NewKustomize(roots.Value()...)
					return getImages(c, kustomize)
				},
			},
			{
				Name: "jaegar",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "mapping",
						Usage: "path to a mapping file",
					},
					&cli.StringFlag{
						Name:  "jaegar-url",
						Usage: "Jaegar URL",
						Value: "http://0.0.0.0:16686",
					},
					&cli.StringFlag{
						Name:     "trace-id",
						Usage:    "trace ID",
						Required: true,
					},

					&cli.StringFlag{
						Name:  "output",
						Value: JSONOutputStr,
						Usage: fmt.Sprintf("specify output format [%s]. Default is '%s'.", JSONOutputStr, JSONOutputStr),
					},
				},
				Action: func(c *cli.Context) error {

					jaegarURL := c.Value("jaegar-url").(string)
					traceID := c.Value("trace-id").(string)

					jaegar := images.NewJaegar(jaegarURL, traceID)
					return getImages(c, jaegar)
				},
			},
		},
	}
}

func getImages(c *cli.Context, provider imageProvider) error { // nolint:funlen

	ctx := context.Background()

	mappingFile := c.Value("mapping").(string)
	output := c.Value("output").(string)
	ghClient := gh.NewClientFromEnv(c.Context)

	all, err := provider.Images()

	if err != nil {
		return err // already wrapped
	}

	var (
		mapper *mapping.Mapper
	)

	if mappingFile != "" {
		mapper, err = mapping.LoadFromFile(mappingFile)
		if err != nil {
			return errors.Wrap(err, "error creating mapper")
		}
	}

	images := []mapping.Image{}

	for _, key := range all.OrderedKeys() {
		bits := strings.Split(key, ":")

		imageName := bits[0]
		version := "unknown"
		if len(bits) == 2 { // nolint: gomnd
			version = bits[1]
		}

		image := mapping.Image{Name: imageName, Version: version, Count: all[key]}

		if mapper != nil {
			clientFound, sonarcloudClient, err := sonarcloud.NewClientFromEnv("https://sonarcloud.io")
			if clientFound && err != nil {
				return errors.Wrapf(err, "error creating sonarcloud client")
			}
			if clientFound {
				if _, err := mapper.Decorate(ctx,ghClient, sonarcloudClient, &image); err != nil {
					return errors.Wrapf(err, "error mapping image '%s' to repo and sonarcloud", bits[0])
				}else {
			if _, err := mapper.Decorate(ctx, ghClient,nil, &image); err != nil {
				return errors.Wrapf(err, "error mapping image '%s' to repo", bits[0])
			}
				}
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
}
