package cmds

import (
	"context"
	"encoding/json"
	"os"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func ListCmd() *cli.Command {
	return &cli.Command{
		Name: "list",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "label",
				Value: "",
				Usage: "specify repository label to predicate on",
			},
			&cli.StringFlag{
				Name:     "owner",
				Value:    "",
				Usage:    "github organisation",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "json",
				Usage: "specify output format [json]. Default is json.",
			},
			&cli.BoolFlag{
				Name:  "omit-archived",
				Value: false,
				Usage: "omit archived repositories",
			},
		},
		Action: func(c *cli.Context) error {

			ctx := context.Background()
			ghClient := gh.NewClient(ctx)

			label := c.Value("label").(string)
			owner := c.Value("owner").(string)
			output := c.Value("output").(string)
			omitArchived := c.Value("omit-archived").(bool)

			repos, err := ghClient.GetReposWithTopic(ctx, owner, label)
			if err != nil {
				return err
			}

			all := []gh.RepositorySlim{}

			for _, repo := range repos {
				if omitArchived && repo.IsArchived {
					continue
				}
				all = append(all, repo)
			}
			switch output {
			case "json":
				b, err := json.Marshal(all)
				if err != nil {
					return errors.Wrap(err, "error marshalling to json")
				}
				os.Stdout.Write(b)
			}
			return nil
		},
	}
}
