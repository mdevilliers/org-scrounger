package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
)

func (c *client) GetUnreleasedCommitsForRepo(ctx context.Context, owner, reponame string) (UnreleasedCommits, RateLimit, error) { // nolint
	ret := UnreleasedCommits{}

	// get last tag - should be a release really but things are a bit weird in this org
	// work through the the commits looking for the oid of the last tag
	var query struct {
		RateLimit  RateLimit `json:"rate_limit"`
		Repository struct {
			Refs struct {
				Nodes []struct {
					Name   githubv4.String `json:"name"`
					Target struct {
						Oid       githubv4.String `json:"oid"`
						CommitURL githubv4.String `json:"commit_url"`
					} `json:"target"`
				} `json:"nodes"`
			} `graphql:"refs(last:1, refPrefix: \"refs/tags/\", orderBy: {field: TAG_COMMIT_DATE, direction: ASC} )" json:"refs"`
			Ref struct {
				Target struct {
					Commit struct {
						History struct {
							Nodes []struct {
								AbbreviatedOid githubv4.String `json:"abbreviated_oid"`
								Oid            githubv4.String `json:"oid"`
								Message        githubv4.String `json:"message"`
								URL            githubv4.String `json:"url"`
							} `json:"nodes"`
						} `json:"history"`
					} `graphql:"... on Commit" json:"commit"`
				} `json:"target"`
			} `graphql:"ref(qualifiedName: \"main\")" json:"ref"`
		} `graphql:"repository(owner:$owner, name:$name)" json:"repository"`
	}
	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(reponame),
	}

	if err := c.graph.Query(ctx, &query, variables); err != nil {
		return ret, query.RateLimit, errors.Wrap(err, "error querying github")
	}
	latestTagOid := "unknown"
	ret.LastTag = Tag{Oid: latestTagOid, Tag: "unknown"}

	if len(query.Repository.Refs.Nodes) == 1 {
		latestTagOid = string(query.Repository.Refs.Nodes[0].Target.Oid)
		commitURL := string(query.Repository.Refs.Nodes[0].Target.CommitURL)
		// if a tagged commit has two parents, trusting the URL commitUrl
		// as the Oid is better. Git is complicated...
		if !strings.Contains(commitURL, latestTagOid) {
			bits := strings.Split(commitURL, "/")
			latestTagOid = bits[len(bits)-1]
		}
		ret.LastTag = Tag{Oid: latestTagOid, Tag: string(query.Repository.Refs.Nodes[0].Name)}
	}

	for _, commit := range query.Repository.Ref.Target.Commit.History.Nodes {
		oid := string(commit.Oid)
		if oid == latestTagOid {
			break
		}
		ret.Commits = append(ret.Commits, Commit{
			Message:        string(commit.Message),
			Oid:            string(commit.Oid),
			AbbreviatedOid: string(commit.AbbreviatedOid),
			URL:            string(commit.URL),
		})
	}

	if len(ret.Commits) == len(query.Repository.Ref.Target.Commit.History.Nodes) {
		ret.Summary = fmt.Sprintf(`%d commits since the last tag.
Are there any tags for the repo?
Or mabe the last tagged commit isn't listed in the commits. Last tag: %s (%s)`,
			len(ret.Commits), ret.LastTag.Tag, ret.LastTag.Oid)
	}
	return ret, query.RateLimit, nil
}
