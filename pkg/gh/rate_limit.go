package gh

import "github.com/shurcooL/githubv4"

// RateLimit contains information from github as to our
// existing rate limit quota for the token
type RateLimit struct {
	Limit     githubv4.Int      `json:"limit"`
	Cost      githubv4.Int      `json:"cost"`
	Remaining githubv4.Int      `json:"remaining"`
	ResetAt   githubv4.DateTime `json:"reset_at"`
}

// Add keeps track of the amalgam of 2 RateLimit tokens
func (rl RateLimit) Add(rl2 RateLimit) RateLimit {
	rl.Cost += rl2.Cost
	rl.Limit = rl2.Limit
	rl.Remaining = rl2.Remaining
	rl.ResetAt = rl2.ResetAt
	return rl
}
