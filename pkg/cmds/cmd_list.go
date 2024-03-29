package cmds

import (
	"context"

	"github.com/mdevilliers/org-scrounger/pkg/cmds/logging"
	"github.com/mdevilliers/org-scrounger/pkg/cmds/output"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/urfave/cli/v3"
)

func listCmd() *cli.Command {
	return &cli.Command{
		Name: "list",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "topic",
				Value: "",
				Usage: "specify repository topic to predicate on",
			},
			&cli.StringFlag{
				Name:     "owner",
				Value:    "",
				Usage:    "github organisation",
				Required: true,
			},
			output.CLIOutputJSONFlag,
			&cli.BoolFlag{
				Name:  "omit-archived",
				Value: false,
				Usage: "omit archived repositories",
			},
			&cli.BoolFlag{
				Name:  "log-rate-limit",
				Value: false,
				Usage: "log the rate limit metrics from github",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {

			ghClient := gh.NewClientFromEnv(ctx)

			topic := c.String("topic")
			owner := c.String("owner")
			omitArchived := c.Bool("omit-archived")
			logRateLimit := c.Bool("log-rate-limit")

			log := logging.GetRateLimitLogger(logRateLimit)

			repos, rateLimit, err := ghClient.GetReposWithTopic(ctx, owner, topic)
			log(rateLimit)
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
			outputter, err := output.GetFromCLIContext(c)
			if err != nil {
				return err
			}
			return outputter(all)
		},
	}
}
