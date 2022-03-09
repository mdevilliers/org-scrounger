package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/mdevilliers/org-scrounger/pkg/cmds"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "team-reporter",
		Action: func(c *cli.Context) error {
			return nil
		},
		Commands: cmds.Commands(),
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
