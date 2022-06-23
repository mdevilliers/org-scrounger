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
		Languages  []string `json:"languages"`
	}

	client struct {
		graph *githubv4.Client
	}
)

// NewClientFromEnv returns a configured client using the env var GITHUB_TOKEN
func NewClientFromEnv(ctx context.Context) *client {
	token := os.Getenv("GITHUB_TOKEN")
	return NewClientFromGithubPAT(ctx, token)
}

//NewClientFromGithubPAT returns a configured client using the supplied Github PAT
func NewClientFromGithubPAT(ctx context.Context, token string) *client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	return &client{
		graph: githubv4.NewClient(httpClient),
	}
}
