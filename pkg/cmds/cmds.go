package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"

	"github.com/Masterminds/sprig"
	"github.com/mdevilliers/org-scrounger/pkg/funcs"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func Commands() []*cli.Command {
	return []*cli.Command{
		GetTeamReportCmd(),
	}
}

func GetTeamReportCmd() *cli.Command {
	return &cli.Command{
		Name: "report",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "label",
				Value: "",
				Usage: "specify repository label to predicate on",
			},
			&cli.StringFlag{
				Name:  "repo",
				Value: "",
				Usage: "specify repository name, required if no label is provided",
			},
			&cli.StringFlag{
				Name:  "owner",
				Value: "",
				Usage: "github organisation",
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "json",
				Usage: "specify output format [html, json]",
			},
			&cli.BoolFlag{
				Name:  "omit-archived",
				Value: false,
				Usage: "omit archived repositories",
			},
			&cli.StringFlag{
				Name:  "template",
				Value: "../../template/index.html",
				Usage: "specify path to go template, required if --output is html",
			},
			&cli.StringSliceFlag{
				Name:    "not-released",
				Aliases: []string{"nr"},
				Usage:   "specify repos that aren't released e.g. a development library or a POC",
			},
		},
		Action: func(c *cli.Context) error {

			ctx := context.Background()
			ghClient := gh.NewClient(ctx)

			label := c.Value("label").(string)
			repo := c.Value("repo").(string)
			owner := c.Value("owner").(string)
			output := c.Value("output").(string)
			templateFile := c.Value("template").(string)
			notReleased := c.Value("not-released").(cli.StringSlice)
			omitArchived := c.Value("omit-archived").(bool)

			if owner == "" {
				return errors.New("Error : supply owner")
			}
			if label == "" {
				if repo == "" {
					return errors.New("Error : supply label or a repo")
				}
			}

			repos := []gh.RepositorySlim{}
			var err error

			if repo != "" {
				repos = append(repos, gh.RepositorySlim{
					Name: repo,
					Url:  fmt.Sprintf("https://github.com/%s/%s", owner, repo),
				})
			} else {
				repos, err = ghClient.GetReposWithTopic(ctx, owner, label)
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
			for _, repo := range repos {
				if omitArchived && repo.IsArchived {
					continue
				}
				reponame := repo.Name
				repoDetails, err := ghClient.GetRepoDetails(ctx, owner, reponame)

				if err != nil {
					return err
				}

				all.Repositories[reponame] = Details{
					Details: repoDetails,
				}

				isReleased := true
				for _, nono := range notReleased.Value() {
					if reponame == nono {
						isReleased = false
					}
				}
				if isReleased {
					unreleasedCommits, err := ghClient.GetUnreleasedCommitsForRepo(ctx, owner, reponame)
					if err != nil {
						return err
					}
					detail := all.Repositories[reponame]
					detail.UnreleasedCommits = unreleasedCommits
					all.Repositories[reponame] = detail
				}
			}

			switch output {
			case "json":
				b, err := json.Marshal(all)
				if err != nil {
					return errors.Wrap(err, "error marshalling to json")
				}
				os.Stdout.Write(b)
			case "html":

				tmpl, err := template.New("index.html").Funcs(funcs.FuncMap()).Funcs(sprig.FuncMap()).ParseFiles(templateFile)

				if err != nil {
					return errors.Wrap(err, "error parsing template")
				}

				if err := tmpl.Execute(os.Stdout, all); err != nil {
					return errors.Wrap(err, "error executing template")
				}
			}
			return nil
		},
	}
}
