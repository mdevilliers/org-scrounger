package main

import (
	"context"
	"os"
	"sort"

	"github.com/mdevilliers/org-scrounger/pkg/cmds"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "scrng",
		Usage: "",
		Action: func(_ context.Context, _ *cli.Command) error {
			return nil
		},
		Commands: cmds.Commands(),
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
