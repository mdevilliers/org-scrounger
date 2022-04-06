package gh

import (
	"context"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type (
	Tag struct {
		Tag string `json:"tag"`
		Oid string `json:"oid"`
	}
	Commit struct {
		Message        string `json:"message"`
		AbbreviatedOid string `json:"abbreviated_oid"`
		Oid            string `json:"oid"`
		Url            string `json:"url"`
	}
	UnreleasedCommits struct {
		Commits []Commit `json:"commits"`
		LastTag Tag      `json:"last_tag"`
		Summary string   `json:"summary"`
	}
	RepositorySlim struct {
		Name       string   `json:"name"`
		Url        string   `json:"url"`
		IsArchived bool     `json:"is_archived"`
		Topics     []string `json:"topics"`
	}
	RateLimit struct {
		Limit     githubv4.Int      `json:"limit"`
		Cost      githubv4.Int      `json:"cost"`
		Remaining githubv4.Int      `json:"remaining"`
		ResetAt   githubv4.DateTime `json:"reset_at"`
	}
	client struct {
		graph *githubv4.Client
	}
)

func NewClient(ctx context.Context) *client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(ctx, src)
	return &client{
		graph: githubv4.NewClient(httpClient),
	}
}

func (rl RateLimit) Add(rl2 RateLimit) RateLimit {
	rl.Cost += rl2.Cost
	rl.Limit = rl2.Limit
	rl.Remaining = rl2.Remaining
	rl.ResetAt = rl2.ResetAt
	return rl
}
