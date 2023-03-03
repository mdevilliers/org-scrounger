package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
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
	return nil, errors.New("unknown output - needs to be template or json")
}

func Templater(wr io.Writer, templateFile string) (func(data any) error, error) {

	_, file := filepath.Split(templateFile)
	tmpl, err := template.New(file).Funcs(FuncMap()).Funcs(sprig.TxtFuncMap()).ParseFiles(templateFile)

	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	return func(data any) error {
		if err := tmpl.Execute(wr, data); err != nil {
			return fmt.Errorf("error executing template: %w", err)
		}
		return nil
	}, nil
}

func JSONer(wr io.Writer) (func(data any) error, error) {
	return func(data any) error {
		b, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("error marshalling to json: %w", err)
		}
		_, err = wr.Write(b)
		return err
	}, nil
}
