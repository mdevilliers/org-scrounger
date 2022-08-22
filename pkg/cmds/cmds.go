package cmds

import (
	"github.com/urfave/cli/v2"
)

const (
	JSONOutputStr = "json"
)

// Commands returns all registered commands
func Commands() []*cli.Command {
	return []*cli.Command{
		reportCmd(),
		listCmd(),
		imagesCmd(),
	}
}
