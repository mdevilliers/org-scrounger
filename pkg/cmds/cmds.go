package cmds

import (
	"encoding/json"
	"os"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
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

func getRateLimitLogger(logRateLimit bool) func(gh.RateLimit) {
	if !logRateLimit {
		return func(gh.RateLimit) {
			// noop
		}
	}
	// TODO : replace with proper logger once stable
	return func(r gh.RateLimit) {
		b, err := json.Marshal(r)
		if err == nil {
			os.Stderr.WriteString(string(b))
			os.Stderr.WriteString("\n")
		}
	}
}
