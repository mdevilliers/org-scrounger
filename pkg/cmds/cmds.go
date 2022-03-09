package cmds

import (
	"context"
	"encoding/json"
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
				Value: "team-ingestion",
				Usage: "specify repository label to predicate on",
			},
			&cli.StringFlag{
				Name:  "owner",
				Value: "Adarga-Ltd",
				Usage: "github organisation",
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "json",
				Usage: "specify output format [html, json]",
			},
			&cli.StringFlag{
				Name:  "template",
				Value: "../../template/index.html",
				Usage: "specify path to go template, required if --output is html",
			},
		},
		Action: func(c *cli.Context) error {

			ctx := context.Background()
			ghClient := gh.NewClient(ctx)

			label := c.Value("label").(string)
			owner := c.Value("owner").(string)
			output := c.Value("output").(string)
			templateFile := c.Value("template").(string)

			repos, err := ghClient.GetReposWithTopic(ctx, owner, label)
			if err != nil {
				return err
			}

			all := map[string]gh.Repository{}

			for _, repo := range repos {
				reponame := repo.Name
				result, err := ghClient.GetRepoDetails(ctx, owner, reponame)

				if err != nil {
					return err
				}
				all[repo.Name] = result
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
