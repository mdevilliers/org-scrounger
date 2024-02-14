package cmds

import (
	"context"
	"fmt"

	"github.com/mdevilliers/org-scrounger/pkg/cmds/output"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
	"github.com/mdevilliers/org-scrounger/pkg/providers/images"
	"github.com/mdevilliers/org-scrounger/pkg/sonarcloud"
	"github.com/urfave/cli/v3"
)

type imageProvider interface {
	Images(ctx context.Context) ([]mapping.Image, error)
}

func imagesCmd() *cli.Command {
	return &cli.Command{
		Name: "images",
		Commands: []*cli.Command{
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
		Action: func(ctx context.Context, c *cli.Command) error {
			paths := c.StringSlice("path")
			deleteCache := c.Bool("delete-cache-on-exit")
			argo := images.NewArgo(deleteCache, paths...)
			return getImages(ctx, c, argo)
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
		Action: func(ctx context.Context, c *cli.Command) error {
			roots := c.StringSlice("root")
			kustomize := images.NewKustomize(roots...)
			return getImages(ctx, c, kustomize)
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
		Action: func(ctx context.Context, c *cli.Command) error {

			jaegarURL := c.String("jaegar-url")
			traceID := c.String("trace-id")

			jaegar := images.NewJaegar(jaegarURL, traceID)
			return getImages(ctx, c, jaegar)
		},
	}
}

func getImages(ctx context.Context, c *cli.Command, provider imageProvider) error {

	mappingFile := c.String("mapping")
	ghClient := gh.NewClientFromEnv(ctx)

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
			return fmt.Errorf("error creating mapper: %w", err)
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
				return fmt.Errorf("error creating sonarcloud client: %w", err)
			}
			if clientFound {
				if _, err := mapper.Decorate(ctx, ghClient, sonarcloudClient, &image); err != nil {
					return fmt.Errorf("error mapping image '%s' to repo and sonarcloud: %w", image.Name, err)
				}
			} else {
				if _, err := mapper.Decorate(ctx, ghClient, nil, &image); err != nil {
					return fmt.Errorf("error mapping image '%s' to repo: %w", image.Name, err)
				}
			}
		}
		if err := outputter(image); err != nil {
			return err
		}
	}
	return nil
}
