package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/Masterminds/sprig"
	"github.com/alitto/pond"
	"github.com/hashicorp/go-multierror"
	"github.com/mdevilliers/org-scrounger/pkg/funcs"
	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/util"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func ReportCmd() *cli.Command {
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
				Name:     "owner",
				Value:    "",
				Usage:    "github organisation",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "json",
				Usage: "specify output format [template, json]. Default is json.",
			},
			&cli.BoolFlag{
				Name:  "omit-archived",
				Value: false,
				Usage: "omit archived repositories",
			},
			&cli.StringFlag{
				Name:  "template-file",
				Value: "../../template/index.html",
				Usage: "specify path to template file. Uses go's template syntax",
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
		Action: func(c *cli.Context) error {

			ctx := context.Background()
			ghClient := gh.NewClient(ctx)

			label := c.Value("label").(string)
			repo := c.Value("repo").(string)
			owner := c.Value("owner").(string)
			output := c.Value("output").(string)
			templateFile := c.Value("template-file").(string)
			notReleased := c.Value("not-released").(cli.StringSlice)
			skipList := c.Value("skip").(cli.StringSlice)
			omitArchived := c.Value("omit-archived").(bool)

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
			var result error

			pool := pond.New(10, 0, pond.MinWorkers(10))

			for _, repo := range repos {

				reponame := repo.Name

				if omitArchived && repo.IsArchived {
					continue
				}

				if util.Contains(skipList.Value(), reponame) {
					continue
				}

				pool.Submit(func() {

					repoDetails, err := ghClient.GetRepoDetails(ctx, owner, reponame)

					if err != nil {
						multierror.Append(result, err)
						return
					}

					all.Repositories[reponame] = Details{
						Details: repoDetails,
					}

					if util.Contains(notReleased.Value(), reponame) {
						unreleasedCommits, err := ghClient.GetUnreleasedCommitsForRepo(ctx, owner, reponame)
						if err != nil {
							multierror.Append(result, err)
							return
						}
						detail := all.Repositories[reponame]
						detail.UnreleasedCommits = unreleasedCommits
						all.Repositories[reponame] = detail
					}
				})
			}

			pool.StopAndWait()

			if result != nil {
				return result
			}

			switch output {
			case "json":
				b, err := json.Marshal(all)
				if err != nil {
					return errors.Wrap(err, "error marshalling to json")
				}
				os.Stdout.Write(b)
			case "template":

				_, file := filepath.Split(templateFile)
				tmpl, err := template.New(file).Funcs(funcs.FuncMap()).Funcs(sprig.FuncMap()).ParseFiles(templateFile)

				if err != nil {
					return errors.Wrap(err, "error parsing template")
				}

				if err := tmpl.Execute(os.Stdout, all); err != nil {
					return errors.Wrap(err, "error executing template")
				}
			default:
				return errors.New("unknown output - needs to be template or json")
			}
			return nil
		},
	}
}
