package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"text/tabwriter"
	"text/template"

	"github.com/kr/text"
	"github.com/russross/blackfriday/v2"
)

var (
	debug      = flag.Bool("debug", false, "debug")
	inputFile  = flag.String("i", "", "input file")
	outputFile = flag.String("o", "", "output file")
	width      = flag.Int("width", 80, "width")

	outputTemplate = template.Must(template.New("output").Parse(`//go:generate go run github.com/twpayne/chezmoi/internal/extract-long-help -i ../REFERENCE.md -o longhelp.go
package cmd

var longHelps = map[string]string{
{{- range $command, $longHelp := . }}
	"{{ $command }}": {{ printf "%q" $longHelp }},
{{- end }}
}
`))
	debugTemplate = template.Must(template.New("debug").Parse(`
{{- range $command, $longHelp := . -}}
# {{ $command }}

{{  $longHelp }}

{{ end -}}
`))

	doubleQuote = []byte("\"")
	indent      = []byte("    ")
	newline     = []byte("\n")
	space       = []byte(" ")
	tab         = []byte("\t")

	renderers = map[blackfriday.NodeType]func(io.Writer, *blackfriday.Node) error{
		blackfriday.Heading:   renderHeading,
		blackfriday.CodeBlock: renderCodeBlock,
		blackfriday.Paragraph: renderParagraph,
		blackfriday.Table:     renderTable,
	}
)

type errUnsupportedNodeType blackfriday.NodeType

func (e errUnsupportedNodeType) Error() string {
	return fmt.Sprintf("unsupported node type: %s", e)
}

func literalText(node *blackfriday.Node) ([]byte, error) {
	b := &bytes.Buffer{}
	var err error
	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch node.Type {
		case blackfriday.Code:
			if _, err = b.Write(doubleQuote); err != nil {
				return blackfriday.Terminate
			}
			if _, err = b.Write(bytes.ReplaceAll(node.Literal, newline, space)); err != nil {
				return blackfriday.Terminate
			}
			if _, err = b.Write(doubleQuote); err != nil {
				return blackfriday.Terminate
			}
		case blackfriday.Text:
			if _, err = b.Write(bytes.ReplaceAll(node.Literal, newline, space)); err != nil {
				return blackfriday.Terminate
			}
		}
		return blackfriday.GoToNext
	})
	return b.Bytes(), err
}

func renderCodeBlock(w io.Writer, codeBlock *blackfriday.Node) error {
	if codeBlock.Type != blackfriday.CodeBlock {
		return errUnsupportedNodeType(codeBlock.Type)
	}
	return renderIndented(w, codeBlock.Literal)
}

func renderHeading(w io.Writer, heading *blackfriday.Node) error {
	if heading.Type != blackfriday.Heading {
		return errUnsupportedNodeType(heading.Type)
	}
	t, err := literalText(heading)
	if err != nil {
		return err
	}
	if _, err := w.Write(t); err != nil {
		return err
	}
	_, err = w.Write(newline)
	return err
}

func renderIndented(w io.Writer, b []byte) error {
	for _, line := range bytes.SplitAfter(b, newline) {
		if _, err := w.Write(indent); err != nil {
			return err
		}
		if _, err := w.Write(line); err != nil {
			return err
		}
	}
	return nil
}

func renderParagraph(w io.Writer, paragraph *blackfriday.Node) error {
	if paragraph.Type != blackfriday.Paragraph {
		return errUnsupportedNodeType(paragraph.Type)
	}
	t, err := literalText(paragraph)
	if err != nil {
		return err
	}
	if _, err := w.Write(text.WrapBytes(t, *width)); err != nil {
		return err
	}
	_, err = w.Write(newline)
	return err
}

func renderTable(w io.Writer, table *blackfriday.Node) error {
	if table.Type != blackfriday.Table {
		return errUnsupportedNodeType(table.Type)
	}
	b := &bytes.Buffer{}
	tw := tabwriter.NewWriter(b, 0, 8, 1, ' ', 0)
	for rowGroup := table.FirstChild; rowGroup != nil; rowGroup = rowGroup.Next {
		if rowGroup.Type != blackfriday.TableHead && rowGroup.Type != blackfriday.TableBody {
			return errUnsupportedNodeType(rowGroup.Type)
		}
		for row := rowGroup.FirstChild; row != nil; row = row.Next {
			if row.Type != blackfriday.TableRow {
				return errUnsupportedNodeType(row.Type)
			}
			for cell := row.FirstChild; cell != nil; cell = cell.Next {
				if cell.Type != blackfriday.TableCell {
					return errUnsupportedNodeType(cell.Type)
				}
				t, err := literalText(cell)
				if err != nil {
					return err
				}
				if _, err := tw.Write(t); err != nil {
					return err
				}
				if _, err := tw.Write(tab); err != nil {
					return err
				}
			}
			if _, err := tw.Write(newline); err != nil {
				return err
			}
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return renderIndented(w, b.Bytes())
}

func renderLongHelp(start, end *blackfriday.Node) (string, error) {
	b := &bytes.Buffer{}
	for node := start; node != end; node = node.Next {
		if node != start {
			if _, err := b.Write(newline); err != nil {
				return "", err
			}
		}
		renderer, ok := renderers[node.Type]
		if !ok {
			return "", errUnsupportedNodeType(node.Type)
		}
		if err := renderer(b, node); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

func extractLongHelps(r io.Reader) (map[string]string, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	longHelps := make(map[string]string)
	state := 0
	var command string
	var start *blackfriday.Node
	b := blackfriday.New(blackfriday.WithExtensions(blackfriday.Tables))
	b.Parse(data).Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if !entering {
			return blackfriday.GoToNext
		}
		switch state {
		case 0:
			if node.Type == blackfriday.Text && string(node.Literal) == "Commands" && node.Parent != nil && node.Parent.Type == blackfriday.Heading && node.Parent.HeadingData.Level == 2 {
				state = 1
			}
		case 1:
			if node.Type == blackfriday.Code && node.Parent != nil && node.Parent.Type == blackfriday.Heading && node.Parent.HeadingData.Level == 3 {
				if start != nil {
					var longHelp string
					longHelp, err = renderLongHelp(start, node)
					if err != nil {
						return blackfriday.Terminate
					}
					longHelps[command] = longHelp
				}
				command, start = string(node.Literal), node.Parent.Next
				state = 2
			}
		case 2:
			if node.Type == blackfriday.Heading {
				if node.HeadingData.Level <= 3 {
					var longHelp string
					longHelp, err = renderLongHelp(start, node)
					if err != nil {
						return blackfriday.Terminate
					}
					longHelps[command] = longHelp
					command, start = "", nil
					state = 1
				}
				if node.HeadingData.Level <= 2 {
					return blackfriday.Terminate
				}
			}
		}
		return blackfriday.GoToNext
	})
	return longHelps, err
}

func run() error {
	flag.Parse()

	var r io.Reader
	if *inputFile == "" {
		r = os.Stdin
	} else {
		fr, err := os.Open(*inputFile)
		if err != nil {
			return err
		}
		defer fr.Close()
		r = fr
	}

	longHelps, err := extractLongHelps(r)
	if err != nil {
		return err
	}

	var w io.Writer
	if *outputFile == "" {
		w = os.Stdout
	} else {
		fw, err := os.Create(*outputFile)
		if err != nil {
			return err
		}
		defer fw.Close()
		w = fw
	}

	if *debug {
		return debugTemplate.ExecuteTemplate(w, "debug", longHelps)
	}

	buf := &bytes.Buffer{}
	if err := outputTemplate.ExecuteTemplate(buf, "output", longHelps); err != nil {
		return err
	}
	cmd := exec.Command("gofmt", "-s")
	cmd.Stdin = buf
	cmd.Stdout = w
	return cmd.Run()
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
