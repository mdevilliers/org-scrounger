package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/mdevilliers/org-scrounger/pkg/funcs"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

const (
	JSONOutputStr = "json"
)

var (
	CLIOutputTemplateJSONFlag = &cli.StringFlag{
		Name:  "output",
		Value: JSONOutputStr,
		Usage: fmt.Sprintf("specify output format [template, %s]. Default is '%s'.", JSONOutputStr, JSONOutputStr),
	}

	CLIOutputJSONFlag = &cli.StringFlag{
		Name:  "output",
		Value: JSONOutputStr,
		Usage: fmt.Sprintf("specify output format [%s]. Default is '%s'.", JSONOutputStr, JSONOutputStr),
	}

	CLITemplateFileFlag = &cli.StringFlag{
		Name:        "template-file",
		DefaultText: "../../template/index.html",
		Usage:       "specify path to template file. Uses go's template syntax",
	}
)

func GetFromCLIContext(ctx *cli.Context) (func(data any) error, error) {
	out := ctx.Value("output").(string)

	if ctx.IsSet("template-file") {
		templateFile := ctx.Value("template-file").(string)
		return Templater(os.Stdout, templateFile)
	}

	if out == JSONOutputStr {
		return JSONer(os.Stdout)
	}
	return NoOp()
}

func Templater(wr io.Writer, templateFile string) (func(data any) error, error) {

	_, file := filepath.Split(templateFile)
	tmpl, err := template.New(file).Funcs(funcs.FuncMap()).Funcs(sprig.TxtFuncMap()).ParseFiles(templateFile)

	if err != nil {
		return nil, errors.Wrap(err, "error parsing template")
	}

	return func(data any) error {
		if err := tmpl.Execute(wr, data); err != nil {
			return errors.Wrap(err, "error executing template")
		}
		return nil
	}, nil
}

func JSONer(wr io.Writer) (func(data any) error, error) {
	return func(data any) error {
		b, err := json.Marshal(data)
		if err != nil {
			return errors.Wrap(err, "error marshalling to json")
		}
		_, err = wr.Write(b)
		return err
	}, nil
}

func NoOp() (func(data any) error, error) {
	return nil, errors.New("unknown output - needs to be template or json")
}
