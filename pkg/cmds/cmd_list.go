package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func listCmd() *cli.Command { // nolint: funlen
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
			&cli.StringFlag{
				Name:  "output",
				Value: JSONOutputStr,
				Usage: fmt.Sprintf("specify output format [%s]. Default is '%s'.", JSONOutputStr, JSONOutputStr),
			},
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
		Action: func(c *cli.Context) error {

			ctx := context.Background()
			ghClient := gh.NewClientFromEnv(ctx)

			topic := c.Value("topic").(string)
			owner := c.Value("owner").(string)
			output := c.Value("output").(string)
			omitArchived := c.Value("omit-archived").(bool)
			logRateLimit := c.Value("log-rate-limit").(bool)

			log := getRateLimitLogger(logRateLimit)

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
			switch output {
			case JSONOutputStr:
				b, err := json.Marshal(all)
				if err != nil {
					return errors.Wrap(err, "error marshalling to json")
				}
				os.Stdout.Write(b)
			default:
				return errors.New("unknown output")
			}

			return nil
		},
	}
}
