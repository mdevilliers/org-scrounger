package gh

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type client struct {
	graph *githubv4.Client
}

func NewClient(ctx context.Context) *client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(ctx, src)
	return &client{
		graph: githubv4.NewClient(httpClient),
	}
}

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
)

func (c *client) GetReposWithTopic(ctx context.Context, owner, topic string) ([]RepositorySlim, error) {

	var query struct {
		Search struct {
			RepositoryCount githubv4.Int `json:"repositoryCount"`
			PageInfo        struct {
				HasNextPage githubv4.Boolean `json:"has_next_page"`
				EndCursor   githubv4.String  `json:"end_cursor"`
			} `json:"page_info"`
			Nodes []struct {
				Repository struct {
					Name             githubv4.String  `json:"name"`
					Url              githubv4.String  `json:"url"`
					IsArchived       githubv4.Boolean `json:"is_archived"`
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
	for {
		if err := c.graph.Query(ctx, &query, variables); err != nil {
			return nil, errors.Wrap(err, "error querying github")
		}

		for _, r := range query.Search.Nodes {
			topics := []string{}
			for _, t := range r.Repository.RepositoryTopics.Nodes {
				topics = append(topics, string(t.Topic.Name))
			}
			slim := RepositorySlim{
				Name:       string(r.Repository.Name),
				Url:        string(r.Repository.Url),
				IsArchived: bool(r.Repository.IsArchived),
				Topics:     topics,
			}
			ret = append(ret, slim)
		}
		if !query.Search.PageInfo.HasNextPage {
			break
		}
		variables["repositoryCursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
	}
	return ret, nil
}

type Repository struct {
	Name             githubv4.String  `json:"name"`
	Url              githubv4.String  `json:"url"`
	IsArchived       githubv4.Boolean `json:"is_archived"`
	RepositoryTopics struct {
		Nodes []struct {
			Topic struct {
				Name githubv4.String `json:"name"`
			} `json:"topic"`
		} `json:"nodes"`
	} `graphql:"repositoryTopics(first:10)" json:"repository_topics"`
	PullRequests struct {
		Nodes []struct {
			Title     githubv4.String   `json:"title"`
			State     githubv4.String   `json:"state"`
			Mergeable githubv4.String   `json:"mergeable"`
			CreatedAt githubv4.DateTime `json:"created_at"`
			Url       githubv4.String   `json:"url"`
			IsDraft   githubv4.Boolean  `json:"is_draft"`
			Author    struct {
				Login githubv4.String `json:"login"`
			} `json:"author"`
			Commits struct {
				Nodes []struct {
					Commit struct {
						Status struct {
							State githubv4.String `json:"state"`
						} `json:"status"`
					} `json:"commit"`
				} `json:"nodes"`
			} `graphql:"commits(last:1)" json:"commits"`
		} `json:"nodes"`
	} `graphql:"pullRequests(last:30, states:[OPEN])" json:"pull_requests"`
	VulnerabilityAlerts `graphql:"vulnerabilityAlerts(first:100, states:[OPEN])" json:"vulnerability_alerts"`
}

type VulnerabilityAlerts struct {
	Edges Edges `json:"edges"`
}

type Edges []struct {
	Node struct {
		CreatedAt                  githubv4.DateTime `json:"created_at"`
		Number                     githubv4.Int      `json:"number"`
		VulnerableManifestFilename githubv4.String   `json:"vulnerable_manifest_filename"`
		VulnerableManifestPath     githubv4.String   `json:"vulnerable_manifest_path"`
		VulnerableRequirements     githubv4.String   `json:"vulnerable_requirements"`
		SecurityVulnerability      struct {
			Severity githubv4.String `json:"severity"`
			Package  struct {
				Name      githubv4.String `json:"name"`
				Ecosystem githubv4.String `json:"ecosystem"`
			} `json:"package"`
			Advisory struct {
				Description githubv4.String `json:"description"`
			} `json:"advisory"`
			FirstPatchedVersion struct {
				Identifier githubv4.String `json:"identifier"`
			} `json:"first_patched_version"`
		} `json:"security_vulnerability"`
	} `json:"node"`
}

func (c *client) GetRepoDetails(ctx context.Context, owner, reponame string) (Repository, error) {

	var query struct {
		Repository `graphql:"repository(owner:$owner, name:$name)" json:"repository"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(reponame),
	}

	if err := c.graph.Query(ctx, &query, variables); err != nil {
		return Repository{}, errors.Wrapf(err, "error querying github for repo details of %s/%s", owner, reponame)
	}
	return query.Repository, nil
}

func (c *client) GetUnreleasedCommitsForRepo(ctx context.Context, owner, reponame string) (UnreleasedCommits, error) {
	ret := UnreleasedCommits{}

	// get last tag - should be a release really but things are a bit weird in this org
	// work through the the commits looking for the oid of the last tag
	var query struct {
		Repository struct {
			Refs struct {
				Nodes []struct {
					Name   githubv4.String `json:"name"`
					Target struct {
						Oid       githubv4.String `json:"oid"`
						CommitUrl githubv4.String `json:"commit_url"`
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
								Url            githubv4.String `json:"url"`
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
		return ret, errors.Wrap(err, "error querying github")
	}

	latestTagOid := "unknown"
	ret.LastTag = Tag{Oid: latestTagOid, Tag: "unknown"}

	if len(query.Repository.Refs.Nodes) == 1 {
		latestTagOid = string(query.Repository.Refs.Nodes[0].Target.Oid)
		commitUrl := string(query.Repository.Refs.Nodes[0].Target.CommitUrl)
		// if a tagged commit has two parents, trusting the URL commitUrl
		// as the Oid is better. Git is complicated...
		if !strings.Contains(commitUrl, latestTagOid) {
			bits := strings.Split(commitUrl, "/")
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
			Url:            string(commit.Url),
		})
	}

	if len(ret.Commits) == len(query.Repository.Ref.Target.Commit.History.Nodes) {
		ret.Summary = fmt.Sprintf("%d commits since the last tag. Are there any tags for the repo? Or mabe the last tagged commit isn't listed in the commits. Last tag: %s (%s) ", len(ret.Commits), ret.LastTag.Tag, ret.LastTag.Oid)

	}
	return ret, nil
}
