package parser

import (
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type MappingRuleSet struct {
	Entries []*Entry `parser:"@@*"`
}

type Entry struct {
	Comment *string  `parser:"@Comment"`
	Field   *Field   `parser:"| @@"`
	Mapping *Mapping `parser:"| @@"`
}

type Field struct {
	Key   string `parser:"@Ident '='"`
	Value *Value `parser:"@@"`
}

type Mapping struct {
	Key    string `parser:"( @Ident"`
	Ignore *bool  `parser:" | @Wildcard )"`
	Value  *Value `parser:"'>' @@"`
}

type Value struct {
	String   *string  `parser:"@String"`
	List     []*Value `parser:" | '[' ( @@ ( ',' @@ )* )? ']'"`
	Wildcard *bool    `parser:" | @Wildcard"`
}

var (
	parser = participle.MustBuild(&MappingRuleSet{},
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

func UnMarshal(filename string, in io.Reader) (*MappingRuleSet, error) {
	o := &MappingRuleSet{}
	err := parse(filename, in, o)
	return o, err
}

func parse(filename string, in io.Reader, out interface{}) error {
	return parser.Parse(filename, in, out)
}
