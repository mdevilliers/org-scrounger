package images

import (
	"context"

	"github.com/mdevilliers/org-scrounger/pkg/jaegar"
	"github.com/mdevilliers/org-scrounger/pkg/mapping"
)

type jaegarProvider struct {
	traceID string
	url     string
}

func NewJaegar(url, traceID string) *jaegarProvider {
	return &jaegarProvider{
		url:     url,
		traceID: traceID,
	}
}

func (j *jaegarProvider) Images(ctx context.Context) ([]mapping.Image, error) {
	all := []mapping.Image{}

	client, err := jaegar.NewClient(j.url)
	if err != nil {
		return nil, err // already wrapped
	}

	trace, err := client.GetTraceByID(ctx, j.traceID)
	if err != nil {
		return nil, err // already wrapped
	}

	for _, d := range trace.Data {
		for _, p := range d.Processes {
			all = append(all, mapping.Image{
				Name: p.ServiceName,
			})
		}
	}

	return all, nil
}
