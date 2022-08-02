package mapping

import (
	"context"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/sonarcloud"
	"golang.org/x/exp/slices"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	imageNamespace      = "image"
	sonarcloudNamespace = "sonarcloud"
)

//counterfeiter:generate . repoGetter
type repoGetter interface {
	GetRepoByName(ctx context.Context, owner, reponame string) (gh.RepositorySlim, gh.RateLimit, error)
}

//counterfeiter:generate . measureGetter
type measureGetter interface {
	GetMeasures(ctx context.Context, componentID string) (*sonarcloud.MeasureResponse, error)
}

type (
	Image struct {
		Name       string             `json:"name"`
		Version    string             `json:"version"`
		Count      int                `json:"count"`
		Repo       *gh.RepositorySlim `json:"repo,omitempty"`
		Sonarcloud *Sonarcloud        `json:"sonarcloud,omitempty"`
	}
	Sonarcloud struct {
		CodeCoverage struct {
			Value float64 `json:"value"`
		} `json:"code_coverage"`
	}
)

func (m *Mapper) Decorate(ctx context.Context, rg repoGetter, mg measureGetter, image *Image) (bool, error) {

	clean := m.cleanImageName(image.Name)
	status, resolved, keys := m.resolve(imageNamespace, clean)
	switch status {
	case ignored:
		return false, nil
	case noMappingFound:
		resolved = clean
	}

	org, reponame := split(resolved, m.defaultOwner)

	repo, _, err := rg.GetRepoByName(ctx, org, reponame)
	if err != nil {
		return false, err
	}
	image.Repo = &repo

	if len(keys) > 0 && mg != nil {
		// look for sonargraph client ID
		for _, k := range keys {
			if strings.HasPrefix(k, sonarcloudNamespace) {
				bits := strings.Split(k, ":")
				sonarcloudKey := bits[1]
				measures, err := mg.GetMeasures(ctx, sonarcloudKey)

				if err != nil {
					// sonarcloud info is optional so don't error
					// TODO : maybe we need to log the negative?
					return false, nil
				}

				result := &Sonarcloud{}
				if len(measures.Measures) > 0 {
					codeCoverage := measures.Measures[0]

					slices.SortFunc(codeCoverage.History, func(a, b sonarcloud.History) bool {
						return a.Time.After(b.Time.Time)
					})

					result.CodeCoverage.Value = measures.Measures[0].History[0].Value
					image.Sonarcloud = result
				}
			}
		}
	}

	return true, nil
}

func split(repo, defaultOwner string) (string, string) {
	bits := strings.Split(repo, "/")
	if len(bits) == 1 {
		return defaultOwner, repo
	}
	return bits[0], bits[1]
}

func (m *Mapper) cleanImageName(name string) string {
	for k := range m.containerRepos {
		if strings.HasPrefix(name, k) {
			return strings.Replace(name, k, "", 1)
		}
	}
	return name
}
