package gh

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
)

func (c *client) GetRepoByName(ctx context.Context, owner, repo string) (RepositorySlim, RateLimit, error) {

	var query struct {
		RateLimit  RateLimit `json:"rate_limit" graphql:"rate_limit"`
		Repository struct {
			Name       githubv4.String  `json:"name" graphql:"name"`
			Url        githubv4.String  `json:"url" graphql:"url"`
			IsArchived githubv4.Boolean `json:"is_archived" graphql:"is_archived"`
			Languages  struct {
				Nodes []struct {
					Name githubv4.String `json:"name" graphql:"name"`
				} `json:"nodes" graphql:"nodes"`
			} `json:"languages" graphql:"languages(first:10)"`
			RepositoryTopics struct {
				Nodes []struct {
					Topic struct {
						Name githubv4.String `json:"name" graphql:"name"`
					} `json:"topic" graphql:"topic"`
				} `json:"nodes" graphql:"nodes"`
			} `graphql:"repositoryTopics(first:10)" json:"repository_topics"`
		} `graphql:"repository( name : $name, owner : $owner )" json:"repository"`
	}

	variables := map[string]interface{}{
		"name":  githubv4.String(repo),
		"owner": githubv4.String(owner),
	}

	if err := c.graph.Query(ctx, &query, variables); err != nil {
		return RepositorySlim{}, RateLimit{}, errors.Wrap(err, "error querying github")
	}
	r := query.Repository

	topics := []string{}
	for _, t := range r.RepositoryTopics.Nodes {
		topics = append(topics, string(t.Topic.Name))
	}
	languages := []string{}
	for _, l := range r.Languages.Nodes {
		languages = append(languages, string(l.Name))
	}
	slim := RepositorySlim{
		Name:       string(r.Name),
		Url:        string(r.Url),
		IsArchived: bool(r.IsArchived),
		Topics:     topics,
		Languages:  languages,
	}
	return slim, query.RateLimit, nil
}
