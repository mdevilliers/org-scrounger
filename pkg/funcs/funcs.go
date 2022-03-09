package funcs

import (
	"html/template"
	"sort"
	"time"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/shurcooL/githubv4"
)

func FuncMap() map[string]interface{} {
	return template.FuncMap{

		// helpers to deal with the githubv4 types
		"github_toString":   func(s githubv4.String) string { return string(s) },
		"github_toDateTime": func(s githubv4.DateTime) time.Time { return s.Time },

		"predicate_severity": func(va gh.VulnerabilityAlerts, severity ...string) gh.VulnerabilityAlerts {
			ret := gh.VulnerabilityAlerts{}
			for i := range va.Edges {
				s := va.Edges[i].Node.SecurityVulnerability.Severity
				for _, sev := range severity {
					if string(s) == sev {
						ret.Edges = append(ret.Edges, va.Edges[i])
						break
					}
				}
			}

			sort.Sort(BySeverity(ret.Edges))
			return ret
		},
	}
}

type BySeverity gh.Edges

func (a BySeverity) Len() int { return len(a) }
func (a BySeverity) Less(i, j int) bool {
	return a[i].Node.SecurityVulnerability.Severity < a[j].Node.SecurityVulnerability.Severity
}
func (a BySeverity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
