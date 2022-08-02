package mapping

import (
	"fmt"

	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
)

type status int

const (
	ok status = iota
	ignored
	noMappingFound
)

func (m *Mapper) expand(rules *parser.MappingRuleSet) error {

	for _, e := range rules.Entries {
		if e.Field != nil {
			if e.Field.Key == "owner" {
				m.defaultOwner = *(e.Field.Value.String)
			}
			if e.Field.Key == "container_repositories" {
				for _, v := range e.Field.Value.List {
					key := *(v.String)
					m.containerRepos[key] = true
				}
			}
		}
		if e.Mapping != nil {
			if e.Mapping.Ignore != nil {
				ignore := *(e.Mapping.Ignore)
				if ignore {
					key := *(e.Mapping.Value.String)
					m.ignore[key] = true
				}
			} else if e.Mapping.Value != nil {
				if e.Mapping.Value.Wildcard != nil { // nolint: gocritic
					key := e.Mapping.Key
					m.static[key] = true
				} else if e.Mapping.Value.String != nil {
					v := *(e.Mapping.Value.String)
					m.reversed[v] = e.Mapping.Key
					m.keyed[e.Mapping.Key] = []string{v}
				} else if len(e.Mapping.Value.List) != 0 {
					all := []string{}
					for _, v := range e.Mapping.Value.List {
						vv := *(v.String)
						m.reversed[vv] = e.Mapping.Key
						all = append(all, vv)
					}
					m.keyed[e.Mapping.Key] = all
				}
			}
		}
	}

	return nil
}
func (m *Mapper) resolve(namespace, name string) (status, string, []string) {

	needle := name
	if namespace != "" {
		needle = fmt.Sprintf("%s:%s", namespace, name)
	}

	_, found := m.ignore[needle]
	if found {
		return ignored, name, m.keyed[needle]
	}
	v, found := m.reversed[needle]
	if found {
		return ok, v, m.keyed[needle]
	}

	// try resolving with no namespace
	if namespace != "" {
		return m.resolve("", name)
	}
	return noMappingFound, name, m.keyed[name]
}
