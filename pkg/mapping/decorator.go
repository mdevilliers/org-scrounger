package mapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/mdevilliers/org-scrounger/pkg/sonarcloud"
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

func (m *Mapper) MapSonarcloudMeta(ctx context.Context, client measureGetter, image *Image) (bool, error) {

	clean := m.cleanImageName(image.Name)

	status, sonarcloudKey := m.resolve(sonarcloudNamespace, clean)
	switch status {
	case noMappingFound:
		return false, nil
	}

	measures, err := client.GetMeasures(ctx, sonarcloudKey)

	if err != nil {
		// sonarcloud info is optional so don't error
		// TODO : maybe we need to log the negative?
		return false, nil
	}

	fmt.Println(measures)
	return true, nil
}

func (m *Mapper) MapGitHubMeta(ctx context.Context, client repoGetter, image *Image) (bool, error) {

	clean := m.cleanImageName(image.Name)
	status, resolved := m.resolve(imageNamespace, clean)
	switch status {
	case ignored:
		return false, nil
	case noMappingFound:
		resolved = clean
	}

	org, reponame := split(resolved, m.defaultOwner)

	repo, _, err := client.GetRepoByName(ctx, org, reponame)
	if err != nil {
		return false, err
	}
	image.Repo = &repo
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
