package cmds

import (
	"context"

	"github.com/mdevilliers/org-scrounger/pkg/cmds/output"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
	"github.com/mdevilliers/org-scrounger/pkg/providers/images"
	"github.com/mdevilliers/org-scrounger/pkg/sonarcloud"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type imageProvider interface {
	Images(ctx context.Context) ([]mapping.Image, error)
}

func imagesCmd() *cli.Command {
	return &cli.Command{
		Name: "images",
		Subcommands: []*cli.Command{
			imagesArgoCommand(),
			imagesJaegarCommand(),
			imagesKustomizeCommand(),
		},
	}
}

func imagesArgoCommand() *cli.Command {
	return &cli.Command{
		Name: "argo",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "path to argo application",
			},
			&cli.StringFlag{
				Name:  "mapping",
				Usage: "path to a mapping file",
			},
			&cli.BoolFlag{
				Name:  "delete-cache-on-exit",
				Usage: "deletes all caches on exit",
			},
			output.CLIOutputJSONFlag,
		},
		Action: func(c *cli.Context) error {
			paths := c.Value("path").(cli.StringSlice)
			deleteCache := c.Value("delete-cache-on-exit").(bool)
			argo := images.NewArgo(deleteCache, paths.Value()...)
			return getImages(c, argo)
		},
	}
}

func imagesKustomizeCommand() *cli.Command {
	return &cli.Command{
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
			output.CLIOutputJSONFlag,
		},
		Action: func(c *cli.Context) error {
			roots := c.Value("root").(cli.StringSlice)
			kustomize := images.NewKustomize(roots.Value()...)
			return getImages(c, kustomize)
		},
	}
}

func imagesJaegarCommand() *cli.Command {
	return &cli.Command{
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
			output.CLIOutputJSONFlag,
		},
		Action: func(c *cli.Context) error {

			jaegarURL := c.Value("jaegar-url").(string)
			traceID := c.Value("trace-id").(string)

			jaegar := images.NewJaegar(jaegarURL, traceID)
			return getImages(c, jaegar)
		},
	}
}

func getImages(c *cli.Context, provider imageProvider) error {

	ctx := context.Background()

	mappingFile := c.Value("mapping").(string)
	ghClient := gh.NewClientFromEnv(c.Context)

	all, err := provider.Images(ctx)

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
		static := mapper.Static()
		all = append(all, static...)
	}

	outputter, err := output.GetFromCLIContext(c)
	if err != nil {
		return err
	}

	for n := range all {

		image := all[n]
		if mapper != nil {
			clientFound, sonarcloudClient, err := sonarcloud.NewClientFromEnv("https://sonarcloud.io")
			if clientFound && err != nil {
				return errors.Wrapf(err, "error creating sonarcloud client")
			}
			if clientFound {
				if _, err := mapper.Decorate(ctx, ghClient, sonarcloudClient, &image); err != nil {
					return errors.Wrapf(err, "error mapping image '%s' to repo and sonarcloud", image.Name)
				}
			} else {
				if _, err := mapper.Decorate(ctx, ghClient, nil, &image); err != nil {
					return errors.Wrapf(err, "error mapping image '%s' to repo", image.Name)
				}
			}
		}
		if err := outputter(image); err != nil {
			return err
		}
	}
	return nil
}
