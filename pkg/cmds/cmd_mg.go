package cmds

import (
	"context"
	"fmt"

	"github.com/mdevilliers/org-scrounger/pkg/cmds/logging"
	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/urfave/cli/v2"
)

func mgCmd() *cli.Command {
	return &cli.Command{
		Name: "mg",
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
			&cli.StringFlag{
				Name:     "commit-message",
				Value:    "",
				Usage:    "commit message",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "branch",
				Value:    "",
				Usage:    "branch",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "debug",
				Usage: "log leve",
			},
			&cli.StringFlag{
				Name:     "script-path",
				Value:    "",
				Usage:    "script to run",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Value: false,
				Usage: "run without pushing changes or creating pull requests.",
			},
		},
		Action: func(c *cli.Context) error {

			ctx := context.Background()
			ghClient := gh.NewClientFromEnv(ctx)

			topic := c.Value("topic").(string)
			owner := c.Value("owner").(string)
			omitArchived := c.Value("omit-archived").(bool)
			logRateLimit := c.Value("log-rate-limit").(bool)

			logLevel := c.Value("log-level").(string)
			branchName := c.Value("branch").(string)
			commitMessage := c.Value("commit-message").(string)
			scriptPath := c.Value("script-path").(string)
			dryRun := c.Value("dry-run").(bool)

			log := logging.GetRateLimitLogger(logRateLimit)

			repos, rateLimit, err := ghClient.GetReposWithTopic(ctx, owner, topic)
			log(rateLimit)
			if err != nil {
				return err
			}

			for _, repo := range repos {
				if omitArchived && repo.IsArchived {
					continue
				}

				args := []string{
					"run", scriptPath,
					"--log-level", logLevel,
					"--git-type", "cmd",
					"--branch", branchName,
					"--commit-message", quote(commitMessage),
					"--repo", fmt.Sprintf("%s/%s", owner, repo.Name),
				}

				if dryRun {
					args = append(args, "--dry-run")
				}
				output, err := exec.GetCommandOutput(".", "multi-gitter", args...)
				fmt.Println(output)
				if err != nil {
					fmt.Println(err)
				}
			}
			return nil
		},
	}
}

func quote(in string) string {
	return fmt.Sprintf("'%s'", in)
}
