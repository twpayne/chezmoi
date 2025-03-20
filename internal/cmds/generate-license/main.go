package main

// FIXME merge this with internal/cmds/generate-helps

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"os"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"

	"github.com/twpayne/chezmoi/v2/assets/chezmoi.io/docs"
)

//go:embed license.go.tmpl
var licenseGoTmpl string

var output = flag.String("o", "", "output")

func run() error {
	flag.Parse()

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(styles.ASCIIStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	licenseMarkdown := strings.TrimPrefix(docs.License, "# License\n\n")
	license, err := renderer.Render(licenseMarkdown)
	if err != nil {
		return err
	}

	licenseGoTemplate, err := template.New("license.go.tmpl").Funcs(template.FuncMap{
		"splitAndQuoteLines": func(s string) string {
			lines := strings.Split(s, "\n")
			quotedLines := make([]string, len(lines))
			for i, line := range lines {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				if i == len(lines)-1 {
					quotedLines[i] = strconv.Quote(line)
				} else {
					quotedLines[i] = strconv.Quote(line + "\n")
				}
			}
			return strings.Join(quotedLines, " +\n")
		},
	}).Parse(licenseGoTmpl)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := licenseGoTemplate.ExecuteTemplate(&buffer, "license.go.tmpl", license); err != nil {
		return err
	}

	outputBytes := buffer.Bytes()
	if formattedOutput, err := format.Source(outputBytes); err == nil {
		outputBytes = formattedOutput
	}

	if *output == "" || *output == "-" {
		if _, err := os.Stdout.Write(outputBytes); err != nil {
			return err
		}
	} else {
		if err := os.WriteFile(*output, outputBytes, 0o666); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
