package mapping

import (
	"context"
	"os"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
	"github.com/pkg/errors"
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

// LoadFromFile returns an initilised Mapping instance or an error
func LoadFromFile(path string, client repoGetter) (*Mapper, error) {

	if path == "" {
		return nil, errors.New("path to mapping file is empty")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening mapping file : %s", path)
	}
	rules, err := parser.UnMarshal(path, file)
	if err != nil {
		return nil, errors.Wrap(err, "error reading mapping file")
	}
	return New(rules, client)
}

// New returns a successfully initilised Mapping instance or an error
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

func (m *Mapper) SonarcloudFromImage(image string) (bool, error) {
	return true, nil
}

// RepositoryFromImage returns whether the repository was found with some metadata or an error
func (m *Mapper) RepositoryFromImage(image string) (bool, gh.RepositorySlim, error) {

	clean := image
	for k := range m.containerRepos {
		if strings.HasPrefix(image, k) {
			clean = strings.Replace(image, k, "", 1)
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
