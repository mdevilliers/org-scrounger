package gh

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
)

func (c *client) GetReposWithTopic(ctx context.Context, owner, topic string) ([]RepositorySlim, RateLimit, error) {

	var query struct {
		RateLimit RateLimit `json:"rate_limit"`
		Search    struct {
			RepositoryCount githubv4.Int `json:"repositoryCount"`
			PageInfo        struct {
				HasNextPage githubv4.Boolean `json:"has_next_page"`
				EndCursor   githubv4.String  `json:"end_cursor"`
			} `json:"page_info"`
			Nodes []struct {
				Repository struct {
					Name       githubv4.String  `json:"name"`
					Url        githubv4.String  `json:"url"`
					IsArchived githubv4.Boolean `json:"is_archived"`
					Languages  struct {
						Nodes []struct {
							Name githubv4.String `json:"name" graphql:"name"`
						} `json:"nodes" graphql:"nodes"`
					} `json:"languages" graphql:"languages(first:10 )"`
					RepositoryTopics struct {
						Nodes []struct {
							Topic struct {
								Name githubv4.String `json:"name"`
							} `json:"topic"`
						} `json:"nodes"`
					} `graphql:"repositoryTopics(first:10)" json:"repository_topics"`
				} `graphql:"... on Repository" json:"repository"`
			} `json:"nodes"`
		} `graphql:"search(query:$query, type: REPOSITORY, first: 100, after: $repositoryCursor)" json:"search"`
	}
	queryStr := fmt.Sprintf("org:%s", owner)
	if topic != "" {
		queryStr = fmt.Sprintf("topic:%s org:%s", topic, owner)
	}

	variables := map[string]interface{}{
		"query":            githubv4.String(queryStr),
		"repositoryCursor": (*githubv4.String)(nil),
	}

	ret := []RepositorySlim{}

	// keep count of the aggregated RateLimit responses
	rl := RateLimit{}

	for {
		if err := c.graph.Query(ctx, &query, variables); err != nil {
			return nil, rl, errors.Wrap(err, "error querying github")
		}
		for _, r := range query.Search.Nodes {
			topics := []string{}
			for _, t := range r.Repository.RepositoryTopics.Nodes {
				topics = append(topics, string(t.Topic.Name))
			}
			languages := []string{}
			for _, l := range r.Repository.Languages.Nodes {
				languages = append(languages, string(l.Name))
			}

			slim := RepositorySlim{
				Name:       string(r.Repository.Name),
				Url:        string(r.Repository.Url),
				IsArchived: bool(r.Repository.IsArchived),
				Topics:     topics,
				Languages:  languages,
			}
			ret = append(ret, slim)
		}

		rl = rl.Add(query.RateLimit)

		if !query.Search.PageInfo.HasNextPage {
			break
		}
		variables["repositoryCursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
	}
	return ret, rl, nil
}
