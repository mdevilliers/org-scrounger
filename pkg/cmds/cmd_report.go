package cmds

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/alitto/pond"
	"github.com/mdevilliers/org-scrounger/pkg/cmds/logging"
	"github.com/mdevilliers/org-scrounger/pkg/cmds/output"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/util"
	"github.com/urfave/cli/v3"
)

func reportCmd() *cli.Command { //nolint: funlen
	return &cli.Command{
		Name: "report",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "topic",
				Value: "",
				Usage: "specify repository topic to predicate on",
			},
			&cli.StringFlag{
				Name:  "repo",
				Value: "",
				Usage: "specify repository name, required if no label is provided",
			},
			&cli.StringFlag{
				Name:     "owner",
				Value:    "",
				Usage:    "github organisation",
				Required: true,
			},
			output.CLIOutputTemplateJSONFlag,
			output.CLITemplateFileFlag,
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
			&cli.StringSliceFlag{
				Name:    "not-released",
				Aliases: []string{"nr"},
				Usage:   "specify repos that aren't released e.g. a development library or a POC",
			},
			&cli.StringSliceFlag{
				Name:    "skip",
				Aliases: []string{"s"},
				Usage:   "specify repos to skip",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {

			ghClient := gh.NewClientFromEnv(ctx)

			topic := c.String("topic")
			repo := c.String("repo")
			owner := c.String("owner")
			notReleased := c.StringSlice("not-released")
			skipList := c.StringSlice("skip")
			omitArchived := c.Bool("omit-archived")
			logRateLimit := c.Bool("log-rate-limit")

			log := logging.GetRateLimitLogger(logRateLimit)

			if topic == "" {
				if repo == "" {
					return errors.New("error : supply topic or a repo")
				}
			}

			repos := []gh.RepositorySlim{}
			var err error
			var rateLimit gh.RateLimit

			if repo != "" {
				repos = append(repos, gh.RepositorySlim{
					Name: repo,
					URL:  fmt.Sprintf("https://github.com/%s/%s", owner, repo),
				})
			} else {
				repos, rateLimit, err = ghClient.GetReposWithTopic(ctx, owner, topic)
				log(rateLimit)
				if err != nil {
					return err
				}
			}

			type (
				Details struct {
					Details           gh.Repository        `json:"details"`
					UnreleasedCommits gh.UnreleasedCommits `json:"unreleased_commits"`
				}
				Data struct {
					Repositories map[string]Details `json:"repositories"`
				}
			)

			all := Data{Repositories: map[string]Details{}}
			allmutex := sync.Mutex{}

			pool := pond.New(5, 0, pond.MinWorkers(3)) //nolint: gomnd
			defer pool.StopAndWait()
			group, ctx := pool.GroupContext(ctx)

			for _, repo := range repos {

				reponame := repo.Name

				if omitArchived && repo.IsArchived {
					continue
				}

				if util.Contains(skipList, reponame) {
					continue
				}

				group.Submit(func() error {

					repoDetails, rateLimit, err := ghClient.GetRepoDetails(ctx, owner, reponame)
					log(rateLimit)
					if err != nil {
						return err
					}
					allmutex.Lock()
					defer allmutex.Unlock()

					all.Repositories[reponame] = Details{
						Details: repoDetails,
					}

					if util.Contains(notReleased, reponame) {
						unreleasedCommits, rateLimit, err := ghClient.GetUnreleasedCommitsForRepo(ctx, owner, reponame)
						log(rateLimit)
						if err != nil {
							return err
						}
						detail := all.Repositories[reponame]
						detail.UnreleasedCommits = unreleasedCommits
						all.Repositories[reponame] = detail
					}
					return nil
				})
			}
			if err := group.Wait(); err != nil {
				return err
			}

			outputter, err := output.GetFromCLIContext(c)
			if err != nil {
				return err
			}
			return outputter(all)
		},
	}
}
