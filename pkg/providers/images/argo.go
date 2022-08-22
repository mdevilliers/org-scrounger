package images
import (
	"context"

	"github.com/mdevilliers/org-scrounger/pkg/util"
)
type argo struct {
	root string
}

func NewArgo(root string) *argo {
	return &argo{
		root: root,
	}
}

func (a *argo) Images(ctx context.Context) (util.Set[string], error) {
	all := util.NewSet[string]()

	return all, nil 
}
