package mapping

import (
	"errors"
	"fmt"
	"os"

	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
)

// Mapper gives high-level access to a parser.MappingRuleSet
type Mapper struct {
	// reversed holds keys indexed by value
	reversed map[string]string
	// keyed holds values for a key
	keyed          map[string][]string
	ignore         map[string]interface{}
	static         map[string]interface{}
	defaultOwner   string
	containerRepos map[string]interface{}
}

// LoadFromFile returns an initilised Mapping instance or an error
func LoadFromFile(path string) (*Mapper, error) {

	if path == "" {
		return nil, errors.New("path to mapping file is empty")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening mapping file: %s :%w", path, err)
	}
	rules, err := parser.UnMarshal(path, file)
	if err != nil {
		return nil, fmt.Errorf("error reading mapping file: %w", err)
	}
	return New(rules), nil
}

// New returns a successfully initilised Mapping instance
func New(rules *parser.MappingRuleSet) *Mapper {
	m := &Mapper{
		reversed:       map[string]string{},
		keyed:          map[string][]string{},
		ignore:         map[string]interface{}{},
		static:         map[string]interface{}{},
		containerRepos: map[string]interface{}{},
	}
	m.expand(rules)
	return m
}

// Static returns the set of statically defined repos
// that wouldn;t be usually discoverable
func (m *Mapper) Static() []Image {
	all := []Image{}

	for k := range m.static {
		all = append(all, Image{Name: k})
	}

	return all
}
