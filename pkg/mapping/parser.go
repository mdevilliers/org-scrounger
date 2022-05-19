package mapping

import (
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type MappingFile struct {
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
	Lexer = lexer.MustSimple([]lexer.SimpleRule{
		lexer.SimpleRule{Name: `Ident`, Pattern: `[a-zA-Z\d][a-zA-Z_\-\/\d]*`},
		lexer.SimpleRule{Name: "String", Pattern: `"[^"]*"`},
		lexer.SimpleRule{Name: "Wildcard", Pattern: `[_]`},
		lexer.SimpleRule{Name: "Punct", Pattern: `\[|]|[-!()+/*=,>]`},
		lexer.SimpleRule{Name: "Comment", Pattern: `#[^\n]+`},
		lexer.SimpleRule{Name: "whitespace", Pattern: `\s+`},
	})
	Parser = participle.MustBuild(&MappingFile{},
		participle.Lexer(
			Lexer,
		),
		participle.Unquote("String"),
	)
)

func Parse(filename string, in io.Reader, out interface{}) error {
	return Parser.Parse(filename, in, out)
}

//type mappings struct{}
