package chezmoi

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/mitchellh/copystructure"
)

// A Template extends text/template.Template with support for directives.
type Template struct {
	name     string
	template *template.Template
	options  TemplateOptions
}

// TemplateOptions are template options that can be set with directives.
type TemplateOptions struct {
	LeftDelimiter  string
	LineEnding     string
	RightDelimiter string
	Options        []string
}

// ParseTemplate parses a template named name from data with the given funcs and
// templateOptions.
func ParseTemplate(name string, data []byte, funcs template.FuncMap, options TemplateOptions) (*Template, error) {
	contents := options.parseAndRemoveDirectives(data)
	tmpl, err := template.New(name).
		Option(options.Options...).
		Delims(options.LeftDelimiter, options.RightDelimiter).
		Funcs(funcs).
		Parse(string(contents))
	if err != nil {
		return nil, err
	}
	return &Template{
		name:     name,
		template: tmpl,
		options:  options,
	}, nil
}

// AddParseTree adds tmpl's parse tree to t.
func (t *Template) AddParseTree(tmpl *Template) (*Template, error) {
	var err error
	t.template, err = t.template.AddParseTree(tmpl.name, tmpl.template.Tree)
	return t, err
}

// Execute executes t with data.
func (t *Template) Execute(data any) ([]byte, error) {
	if data != nil {
		// Make a deep copy of data, in case any template functions modify it.
		var err error
		data, err = copystructure.Copy(data)
		if err != nil {
			return nil, err
		}
	}

	var builder strings.Builder
	if err := t.template.ExecuteTemplate(&builder, t.name, data); err != nil {
		return nil, err
	}
	return []byte(replaceLineEndings(builder.String(), t.options.LineEnding)), nil
}

// parseAndRemoveDirectives updates o by parsing all template directives in data
// and returns data with the lines containing directives removed. The lines are
// removed so that any delimiters do not break template parsing.
func (o *TemplateOptions) parseAndRemoveDirectives(data []byte) []byte {
	directiveMatches := templateDirectiveRx.FindAllSubmatchIndex(data, -1)
	if directiveMatches == nil {
		return data
	}

	// Parse options from directives.
	for _, directiveMatch := range directiveMatches {
		keyValuePairMatches := templateDirectiveKeyValuePairRx.FindAllSubmatch(data[directiveMatch[2]:directiveMatch[3]], -1)
		for _, keyValuePairMatch := range keyValuePairMatches {
			key := string(keyValuePairMatch[1])
			value := maybeUnquote(string(keyValuePairMatch[2]))
			switch key {
			case "left-delimiter":
				o.LeftDelimiter = value
			case "line-ending", "line-endings":
				switch string(keyValuePairMatch[2]) {
				case "crlf":
					o.LineEnding = "\r\n"
				case "lf":
					o.LineEnding = "\n"
				case "native":
					o.LineEnding = nativeLineEnding
				default:
					o.LineEnding = value
				}
			case "right-delimiter":
				o.RightDelimiter = value
			case "missing-key":
				o.Options = append(o.Options, "missingkey="+value)
			}
		}
	}

	return removeMatches(data, directiveMatches)
}

// removeMatches returns data with matchesIndexes removed.
func removeMatches(data []byte, matchesIndexes [][]int) []byte {
	slices := make([][]byte, len(matchesIndexes)+1)
	slices[0] = data[:matchesIndexes[0][0]]
	for i, matchIndexes := range matchesIndexes[1:] {
		slices[i+1] = data[matchesIndexes[i][1]:matchIndexes[0]]
	}
	slices[len(matchesIndexes)] = data[matchesIndexes[len(matchesIndexes)-1][1]:]
	return bytes.Join(slices, nil)
}

// replaceLineEndings replaces all line endings in s with lineEnding. If
// lineEnding is empty it returns s unchanged.
func replaceLineEndings(s, lineEnding string) string {
	if lineEnding == "" {
		return s
	}
	return lineEndingRx.ReplaceAllString(s, lineEnding)
}
