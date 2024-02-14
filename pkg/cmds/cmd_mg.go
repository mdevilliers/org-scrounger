package cmds

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/cmds/logging"
	"github.com/mdevilliers/org-scrounger/pkg/exec"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/urfave/cli/v3"
)

func mgCmd() *cli.Command { //nolint:funlen
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
			&cli.StringFlag{
				Name:  "lang",
				Value: "",
				Usage: "languge selector e.g Go",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {

			ghClient := gh.NewClientFromEnv(ctx)

			topic := c.String("topic")
			owner := c.String("owner")
			omitArchived := c.Bool("omit-archived")
			logRateLimit := c.Bool("log-rate-limit")

			logLevel := c.String("log-level")
			branchName := c.String("branch")

			commitMessage := c.String("commit-message")

			scriptPath := c.String("script-path")
			dryRun := c.Bool("dry-run")
			lang := c.String("lang")
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
				if lang != "" { // naive predicate for languages used in the repo
					if !strings.EqualFold(repo.Languages.Top(), lang) {
						continue
					}
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

				scopedEnvVars := map[string]string{
					"LANGUAGES":        strings.Join(repo.Languages.All(), ","),
					"LANGUAGE_PRIMARY": repo.Languages.Top(),
					"TOPICS":           strings.Join(repo.Topics, ","),
				}

				for k, v := range scopedEnvVars {
					os.Setenv(k, v)
				}
				output, err := exec.GetCommandOutput(".", "multi-gitter", args...)
				fmt.Println(output)
				if err != nil {
					// don;t return on first error
					fmt.Println(err)
				}

				for k := range scopedEnvVars {
					os.Unsetenv(k)
				}

			}
			return nil
		},
	}
}

func quote(in string) string {
	return fmt.Sprintf("\"%s\"", in)
}
