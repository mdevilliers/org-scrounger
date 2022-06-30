package mapping

import (
	"strings"

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
				if e.Mapping.Value.Wildcard != nil {
					key := e.Mapping.Key
					m.static[key] = true
				} else if e.Mapping.Value.String != nil {
					key := *(e.Mapping.Value.String)
					m.containers[key] = e.Mapping.Key
				} else if len(e.Mapping.Value.List) != 0 {
					for _, v := range e.Mapping.Value.List {
						key := *(v.String)
						m.containers[key] = e.Mapping.Key
					}
				}
			}
		}
	}
	return nil
}
func (m *Mapper) resolve(container string) (status, string, string) {

	_, found := m.ignore[container]
	if found {
		return ignored, m.defaultOwner, container
	}
	v, found := m.containers[container]
	if found {
		owner, c := split(v, m.defaultOwner)
		return ok, owner, c
	}

	return noMappingFound, m.defaultOwner, container
}

func split(repo, defaultOwner string) (string, string) {
	bits := strings.Split(repo, "/")
	if len(bits) == 1 {
		return defaultOwner, repo
	}
	return bits[0], bits[1]
}
