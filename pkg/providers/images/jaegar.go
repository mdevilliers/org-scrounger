package images

import (
	"github.com/mdevilliers/org-scrounger/pkg/jaegar"
	"github.com/mdevilliers/org-scrounger/pkg/util"
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

func (j *jaegarProvider) Images() (util.Set[string], error) {
	all := util.NewSet[string]()

	client, err := jaegar.NewClient(j.url)
	if err != nil {
		return all, err // already wrapped
	}

	trace, err := client.GetTraceByID(j.traceID)
	if err != nil {
		return all, err // already wrapped
	}

	for _, d := range trace.Data {
		for _, p := range d.Processes {
			all.Add(p.ServiceName)
		}
	}

	return all, nil
}
