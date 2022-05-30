package mapping

import (
	"context"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type Mapper struct {
	client repoGetter
}

//counterfeiter:generate . repoGetter
type repoGetter interface {
	GetRepoDetails(ctx context.Context, owner, reponame string) (gh.Repository, gh.RateLimit, error)
}

func New(rules *parser.MappingRuleSet, client repoGetter) (*Mapper, error) {
	m := &Mapper{client: client}
	err := m.Expand(rules)
	return m, err
}

func (m *Mapper) RepositoryFromContainer(container string) (bool, gh.Repository, error) {

	status, org, reponame := m.Resolve(container)
	switch status {
	case Ignore:
		return false, gh.Repository{}, nil
	case NoMappingFound:
		reponame = container
	}

	repo, _, err := m.client.GetRepoDetails(context.Background(), org, reponame)
	if err != nil {
		return false, gh.Repository{}, err
	}
	return true, repo, nil
}
