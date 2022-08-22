package logging

import (
	"io"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitNoOp() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: io.Discard})
}

func GetRateLimitLogger(logRateLimit bool) func(gh.RateLimit) {
	if !logRateLimit {
		return func(gh.RateLimit) {
			// noop
		}
	}
	return func(r gh.RateLimit) {
		log.Info().Interface("rate-limit", r).Send()
	}
}
