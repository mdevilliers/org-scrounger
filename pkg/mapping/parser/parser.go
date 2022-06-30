package parser

import (
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// MappingRuleSet contains a set of Entries
type MappingRuleSet struct {
	Entries []*Entry `parser:"@@*"`
}

// Entry can either be a comment, field (assignment) or a mapping
type Entry struct {
	Comment *string  `parser:"@Comment"`
	Field   *Field   `parser:"| @@"`
	Mapping *Mapping `parser:"| @@"`
}

// Field represents an assigned variable
type Field struct {
	Key   string `parser:"@Ident '='"`
	Value *Value `parser:"@@"`
}

// Mapping represents a relationship between the
// left and right values
type Mapping struct {
	Key    string `parser:"( @Ident"`
	Ignore *bool  `parser:" | @Wildcard )"`
	Value  *Value `parser:"'>' @@"`
}

// Value can either be a string or a list of values, or a wildcard
type Value struct {
	String   *string  `parser:"@String"`
	List     []*Value `parser:" | '[' ( @@ ( ',' @@ )* )? ']'"`
	Wildcard *bool    `parser:" | @Wildcard"`
}

var (
	parser = participle.MustBuild[MappingRuleSet](
		participle.Lexer(
			lexer.MustSimple([]lexer.SimpleRule{
				lexer.SimpleRule{Name: `Ident`, Pattern: `[a-zA-Z\d][a-zA-Z_\-\/\d]*`},
				lexer.SimpleRule{Name: "String", Pattern: `"[^"]*"`},
				lexer.SimpleRule{Name: "Wildcard", Pattern: `[_]`},
				lexer.SimpleRule{Name: "Punct", Pattern: `\[|]|[-!()+/*=,>]`},
				lexer.SimpleRule{Name: "Comment", Pattern: `#[^\n]+`},
				lexer.SimpleRule{Name: "whitespace", Pattern: `\s+`},
			}),
		),
		participle.Unquote("String"),
	)
)

// UnMarshal returns a parsed representation of the rules or an error
func UnMarshal(filename string, in io.Reader) (*MappingRuleSet, error) {
	return parser.Parse(filename, in)
}
