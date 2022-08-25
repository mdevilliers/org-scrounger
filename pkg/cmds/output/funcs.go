package output

import (
	"text/template"
	"time"

	"github.com/mdevilliers/org-scrounger/pkg/gh"
	"github.com/shurcooL/githubv4"
	"golang.org/x/exp/slices"
)

// FuncMap returns a map of registered templating helpers.
func FuncMap() map[string]interface{} {
	return template.FuncMap{

		// helpers to deal with the githubv4 types
		"github_toString":    func(s githubv4.String) string { return string(s) },
		"github_toDateTime":  func(s githubv4.DateTime) time.Time { return s.Time },
		"predicate_severity": PredicateOnSeverity,
	}
}

// PredicateOnSeverity filters on a Severity
func PredicateOnSeverity(va gh.VulnerabilityAlerts, severity ...string) gh.VulnerabilityAlerts {

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
	slices.SortFunc(ret.Edges, func(a, b gh.VulnerabilityAlertsEdge) bool {
		return a.Node.SecurityVulnerability.Severity < b.Node.SecurityVulnerability.Severity
	})
	return ret
}
