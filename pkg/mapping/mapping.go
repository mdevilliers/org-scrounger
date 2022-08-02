package mapping

import (
	"os"

	"github.com/mdevilliers/org-scrounger/pkg/mapping/parser"
	"github.com/pkg/errors"
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
		return nil, errors.Wrapf(err, "error opening mapping file : %s", path)
	}
	rules, err := parser.UnMarshal(path, file)
	if err != nil {
		return nil, errors.Wrap(err, "error reading mapping file")
	}
	return New(rules)
}

// New returns a successfully initilised Mapping instance or an error
func New(rules *parser.MappingRuleSet) (*Mapper, error) {
	m := &Mapper{
		reversed:       map[string]string{},
		keyed:          map[string][]string{},
		ignore:         map[string]interface{}{},
		static:         map[string]interface{}{},
		containerRepos: map[string]interface{}{},
	}
	return m, m.expand(rules)
}
