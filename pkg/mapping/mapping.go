package mapping

import (
	"context"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

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

func New(rules *parser.MappingRuleSet, client repoGetter) (*Mapper, error) {
	m := &Mapper{
		client:         client,
		containers:     map[string]string{},
		ignore:         map[string]interface{}{},
		static:         map[string]interface{}{},
		containerRepos: map[string]interface{}{},
	}
	err := m.Expand(rules)
	return m, err
}

func (m *Mapper) RepositoryFromContainer(container string) (bool, gh.RepositorySlim, error) {
	clean := container
	for k := range m.containerRepos {
		if strings.HasPrefix(container, k) {
			clean = strings.Replace(container, k, "", 1)
		}
	}
	status, org, reponame := m.Resolve(clean)
	switch status {
	case Ignore:
		return false, gh.RepositorySlim{}, nil
	case NoMappingFound:
		reponame = clean
	}
	// TODO : propogate context correctly
	repo, _, err := m.client.GetRepoByName(context.Background(), org, reponame)
	if err != nil {
		return false, gh.RepositorySlim{}, err
	}
	return true, repo, nil
}
