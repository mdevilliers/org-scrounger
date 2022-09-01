package images

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/v2/util/io/path"
	"github.com/argoproj/argo-cd/v2/util/helm"
	"github.com/mdevilliers/org-scrounger/pkg/util"
)

type argoProvider struct {
	paths []string
}

func NewArgo(paths ...string) *argoProvider {
	return &argoProvider{
		paths: paths,
	}
}

func (a *argoProvider) Images(ctx context.Context) (util.Set[string], error) {
	all := util.NewSet[string]()

	/*dir, err := ioutil.TempDir(".", "prefix")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
*/
dir := "/Users/markdevilliers/Desktop/git/adarga-ltd/helm-charts"
	repo := helm.HelmRepository{
		Name: "helm-charts",
		Repo: "https://github.com/Adarga-Ltd/helm-charts",
	}
	passCredentials := false
	proxy := ""
	helmApp, err := helm.NewHelmApp(dir, []helm.HelmRepository{repo}, false, "v3", proxy, passCredentials)

	if err != nil {
		return nil, err
	}

	fmt.Println(helmApp)
	valuesPath := dir + "/charts/bench-micro-ui/values.yaml" 
	opts := &helm.TemplateOpts{
		Name: "test",
		Set: map[string]string{
			"service.type": "LoadBalancer",
			"service.port": "1234",
		},
		SetString: map[string]string{
			"service.annotations.prometheus\\.io/scrape": "true",
		},
		Values: []path.ResolvedFilePath{path.ResolvedFilePath(valuesPath)},
	}

out, err := helmApp.Template(opts)
	if err != nil {
		return nil, err
	}

fmt.Println(out)

	return all, nil

}
