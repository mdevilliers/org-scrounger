package gh

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
)

type Repository struct {
	Name       githubv4.String  `json:"name"`
	URL        githubv4.String  `json:"url"`
	IsArchived githubv4.Boolean `json:"is_archived"`
	Languages  struct {
		Edges []struct {
			Size githubv4.Int `json:"size"`
		} `graphql:"edges" json:"edges"`
		Nodes []struct {
			Name githubv4.String `json:"name"`
		} `json:"nodes"`
	} `json:"languages" graphql:"languages(first:10)"`
	Ref struct {
		Target struct {
			Commit struct {
				Message           githubv4.String `json:"message"`
				StatusCheckRollup struct {
					State    githubv4.String `json:"state"`
					Contexts struct {
						Nodes []struct {
							StatusContext `graphql:"... on StatusContext" json:"status_context,omitempty"`
							CheckRun      `graphql:"... on CheckRun" json:"check_run,omitempty"`
						} `json:"nodes"`
					} `json:"contexts" graphql:"contexts(first:20)"`
				} `json:"statusCheckRollup"`
			} `graphql:"... on Commit" json:"commit"`
		} `json:"target"`
	} `graphql:"ref(qualifiedName: \"main\" )" json:"ref"`
	RepositoryTopics struct {
		Nodes []struct {
			Topic struct {
				Name githubv4.String `json:"name"`
			} `json:"topic"`
		} `json:"nodes"`
	} `graphql:"repositoryTopics(first:10)" json:"repository_topics"`
	PullRequests        `graphql:"pullRequests(last:30, states:[OPEN])" json:"pull_requests"`
	VulnerabilityAlerts `graphql:"vulnerabilityAlerts(first:100, states:[OPEN])" json:"vulnerability_alerts"`
}

type CheckRun struct {
	Name       githubv4.String `json:"name,omitempty"`
	Summary    githubv4.String `json:"summary,omitempty"`
	Text       githubv4.String `json:"text,omitempty"`
	Title      githubv4.String `json:"title,omitempty"`
	Status     githubv4.String `json:"status,omitempty"`
	Conclusion githubv4.String `json:"conclusion,omitempty"`
	CheckSuite *struct {
		WorkflowRun *struct {
			Workflow *struct {
				Name githubv4.String `json:"name,omitempty"`
			} `json:"workflow,omitempty"`
		} `json:"workflow_run,omitempty"`
	} `json:"check_suite,omitempty"`
}

func (c CheckRun) IsEmpty() bool {
	return c.Name == ""
}

type StatusContext struct {
	Context githubv4.String `json:"context,omitempty"`
	State   githubv4.String `json:"state,omitempty"`
}

func (s StatusContext) IsEmpty() bool {
	return s.State == ""
}

func (r Repository) IsMainGreen() bool {
	return string(r.Ref.Target.Commit.StatusCheckRollup.State) == "SUCCESS"
}

type PullRequests struct {
	Nodes []PullRequest `json:"nodes"`
}

type PullRequest struct {
	Title      githubv4.String   `json:"title"`
	State      githubv4.String   `json:"state"`
	Mergeable  githubv4.String   `json:"mergeable"`
	CreatedAt  githubv4.DateTime `json:"created_at"`
	URL        githubv4.String   `json:"url"`
	IsDraft    githubv4.Boolean  `json:"is_draft"`
	Repository struct {
		Name githubv4.String `json:"name"`
	} `json:"repository"`
	Author struct {
		Login githubv4.String `json:"login"`
	} `json:"author"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup struct {
					State githubv4.String `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `graphql:"commits(last:1)" json:"commits"`
}

func (p PullRequest) IsMergable() bool {
	return p.Mergeable == "MERGEABLE"
}
func (p PullRequest) LastCommitBuilds() bool {
	return p.Commits.Nodes[0].Commit.StatusCheckRollup.State == "SUCCESS"
}

type VulnerabilityAlerts struct {
	Edges []VulnerabilityAlertsEdge `json:"edges"`
}

type VulnerabilityAlertsEdge struct {
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

func (c *client) GetRepoDetails(ctx context.Context, owner, reponame string) (Repository, RateLimit, error) {

	var query struct {
		Repository `graphql:"repository(owner:$owner, name:$name)" json:"repository"`
		RateLimit  RateLimit `json:"rate_limit"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(reponame),
	}

	if err := c.graph.Query(ctx, &query, variables); err != nil {
		return Repository{}, query.RateLimit, errors.Wrapf(err, "error querying repo details of %s/%s", owner, reponame)
	}
	return query.Repository, query.RateLimit, nil

}
