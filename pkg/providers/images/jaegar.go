package images

import (
	"github.com/mdevilliers/org-scrounger/pkg/util"
)

type jaegar struct {
	traceID string
	url     string
}

func NewJaegar(url, traceID string) *jaegar {
	return &jaegar{
		url:     url,
		traceID: traceID,
	}
}

func (j *jaegar) Images() (util.Set[string], error) {
	all := util.NewSet[string]()
	return all, nil
}
