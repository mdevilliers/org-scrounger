package mapping

import (
	"context"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	imageNamespace = "image"
)

// Mapper gives high-level access to a parser.MappingRuleSet
type Mapper struct {
	client         repoGetter
	containers     map[string]string
	ignore         map[string]interface{}
	static         map[string]interface{}
	defaultOwner   string
	containerRepos map[string]interface{}
}

//counterfeiter:generate . repoGetter
type repoGetter interface {
	GetRepoByName(ctx context.Context, owner, reponame string) (gh.RepositorySlim, gh.RateLimit, error)
}

// New returns a successfully MappingRuleSet or an error
func New(rules *parser.MappingRuleSet, client repoGetter) (*Mapper, error) {
	m := &Mapper{
		client:         client,
		containers:     map[string]string{},
		ignore:         map[string]interface{}{},
		static:         map[string]interface{}{},
		containerRepos: map[string]interface{}{},
	}
	err := m.expand(rules)
	return m, err
}

// RepositoryFromImage returns whether the repository was found with some metadata or an error
func (m *Mapper) RepositoryFromImage(container string) (bool, gh.RepositorySlim, error) {

	clean := container
	for k := range m.containerRepos {
		if strings.HasPrefix(container, k) {
			clean = strings.Replace(container, k, "", 1)
		}
	}
	status, org, reponame := m.resolve(imageNamespace, clean)
	switch status {
	case ignored:
		return false, gh.RepositorySlim{}, nil
	case noMappingFound:
		reponame = clean
	}
	// TODO : propagate context correctly
	repo, _, err := m.client.GetRepoByName(context.Background(), org, reponame)
	if err != nil {
		return false, gh.RepositorySlim{}, err
	}
	return true, repo, nil
}
