package gh

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
)

func (c *client) GetRepoByName(ctx context.Context, owner, repo string) (RepositorySlim, RateLimit, error) {

	var query struct {
		RateLimit  RateLimit `json:"rate_limit"`
		Repository struct {
			Name       githubv4.String  `json:"name"`
			URL        githubv4.String  `json:"url"`
			IsArchived githubv4.Boolean `json:"is_archived"`
			Languages  struct {
				Edges []struct {
					Size githubv4.Int `json:"size"`
				} `json:"edges"`
				Nodes []struct {
					Name githubv4.String `json:"name"`
				} `json:"nodes"`
			} `json:"languages" graphql:"languages(first:10)"`
			RepositoryTopics struct {
				Nodes []struct {
					Topic struct {
						Name githubv4.String `json:"name"`
					} `json:"topic"`
				} `json:"nodes"`
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
	languages := map[string]int{}
	for e, l := range r.Languages.Nodes {
		languages[string(l.Name)] = int(r.Languages.Edges[e].Size)
	}

	slim := RepositorySlim{
		Name:       string(r.Name),
		URL:        string(r.URL),
		IsArchived: bool(r.IsArchived),
		Topics:     topics,
		Languages:  languages,
	}
	return slim, query.RateLimit, nil
}
