package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"

	"github.com/twpayne/chezmoi/v2/assets/chezmoi.io/docs/reference/commands"
	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

//go:embed helps.go.tmpl
var helpsGoTmpl string

type help struct {
	LongHelp   string
	Example    string
	LongFlags  chezmoiset.Set[string]
	ShortFlags chezmoiset.Set[string]
}

var (
	unwantedEscapeSequenceRx = regexp.MustCompile(`\x1b\[[01]m`)
	linkRx                   = regexp.MustCompile(`(?m)\[(.*?)]\[(.*?)]`)
	helpFlagsRx              = regexp.MustCompile("^### (?:`-([0-9A-Za-z])`, )?`--([-0-9A-Za-z]+)`")
	trailingSpaceRx          = regexp.MustCompile(` +\n`)

	output = flag.String("o", "", "output")
)

func run() error {
	flag.Parse()

	dirEntries, err := commands.FS.ReadDir(".")
	if err != nil {
		return err
	}

	longHelpStyleConfig := styles.ASCIIStyleConfig
	longHelpStyleConfig.Code.StylePrimitive.BlockPrefix = ""
	longHelpStyleConfig.Code.StylePrimitive.BlockSuffix = ""
	longHelpStyleConfig.Emph.BlockPrefix = ""
	longHelpStyleConfig.Emph.BlockSuffix = ""
	longHelpStyleConfig.H2.Prefix = ""
	longHelpTermRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(longHelpStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	exampleStyleConfig := styles.ASCIIStyleConfig
	exampleStyleConfig.Code.StylePrimitive.BlockPrefix = ""
	exampleStyleConfig.Code.StylePrimitive.BlockSuffix = ""
	exampleStyleConfig.Document.Margin = nil
	exampleTermRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(exampleStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	helps := make(map[string]*help)
	for _, dirEntry := range dirEntries {
		if dirEntry.Name() == "index.md" {
			continue
		}
		command := strings.TrimSuffix(dirEntry.Name(), ".md")
		data, err := commands.FS.ReadFile(dirEntry.Name())
		if err != nil {
			return err
		}
		help, err := extractHelp(command, data, longHelpTermRenderer, exampleTermRenderer)
		if err != nil {
			return err
		}
		helps[command] = help
	}

	funcMap := template.FuncMap{
		"quoteElementsOnePerLine": func(s chezmoiset.Set[string]) string {
			if s.IsEmpty() {
				return ""
			}
			elements := s.Elements()
			sort.Strings(elements)
			quotedElementLines := make([]string, len(elements))
			for i, element := range elements {
				quotedElementLines[i] = "\n" + strconv.Quote(element) + ","
			}
			return strings.Join(quotedElementLines, "") + "\n"
		},
		"splitAndQuoteLines": func(s string) string {
			lines := strings.Split(s, "\n")
			quotedLines := make([]string, len(lines))
			for i, line := range lines {
				if i == len(lines)-1 {
					quotedLines[i] = strconv.Quote(line)
				} else {
					quotedLines[i] = strconv.Quote(line + "\n")
				}
			}
			return strings.Join(quotedLines, " +\n")
		},
	}
	helpsGoTemplate, err := template.New("helps.go.tmpl").Funcs(funcMap).Parse(helpsGoTmpl)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := helpsGoTemplate.ExecuteTemplate(&buffer, "helps.go.tmpl", helps); err != nil {
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

// extractHelp returns the helps parse from r.
func extractHelp(command string, data []byte, longHelpTermRenderer, exampleTermRenderer *glamour.TermRenderer) (*help, error) {
	type stateType int
	const (
		stateReadTitle stateType = iota
		stateInLongHelp
		stateInOptions
		stateInExamples
		stateInNotes
		stateInAdmonition
		stateInUnknownSection
	)

	state := stateReadTitle
	var longHelpLines []string
	var exampleLines []string
	longFlags := chezmoiset.New[string]()
	shortFlags := chezmoiset.New[string]()

	stateChange := func(line string, state *stateType) bool {
		switch {
		case line == "## Flags" || line == "## Common flags":
			*state = stateInOptions
			return true
		case line == "## Examples":
			*state = stateInExamples
			return true
		case line == "## Notes":
			*state = stateInNotes
			return true
		case strings.HasPrefix(line, "## "):
			*state = stateInUnknownSection
			return true
		}
		return false
	}

	for _, line := range strings.Split(string(data), "\n") {
		switch state {
		case stateReadTitle:
			titleRx, err := regexp.Compile("# `" + command + "`")
			if err != nil {
				return nil, err
			}
			if titleRx.MatchString(line) {
				state = stateInLongHelp
			} else {
				return nil, fmt.Errorf("expected title for '%s'", command)
			}
		case stateInLongHelp:
			switch {
			case stateChange(line, &state):
				break
			case strings.HasPrefix(line, "!!! "):
				state = stateInAdmonition
			default:
				longHelpLines = append(longHelpLines, line)
			}
		case stateInExamples:
			if !stateChange(line, &state) {
				exampleLines = append(exampleLines, line)
			}
		case stateInOptions:
			if !stateChange(line, &state) {
				if m := helpFlagsRx.FindStringSubmatch(line); m != nil {
					if m[1] != "" {
						shortFlags.Add(m[1])
					}
					longFlags.Add(m[2])
				}
			}
		default:
			stateChange(line, &state)
		}
	}

	longHelp, err := renderLines(longHelpLines, longHelpTermRenderer)
	if err != nil {
		return nil, err
	}
	example, err := renderLines(exampleLines, exampleTermRenderer)
	if err != nil {
		return nil, err
	}
	return &help{
		LongHelp:   "Description:\n" + longHelp,
		Example:    example,
		LongFlags:  longFlags,
		ShortFlags: shortFlags,
	}, nil
}

// renderLines renders lines, trimming extraneous whitespace.
func renderLines(lines []string, termRenderer *glamour.TermRenderer) (string, error) {
	in := strings.Join(lines, "\n")
	// Replace links with their anchor text only. This should be possible by
	// using Conceal in the style, but this currently has no effect. See
	// https://github.com/charmbracelet/glamour/issues/389.
	in = linkRx.ReplaceAllString(in, "$1")
	// For some reason, the above regular expression does not work if the link
	// spans multiple lines. Fix this with a horrible hack for upgrade.md.
	in = strings.ReplaceAll(in, "[rate\nlimiting][rate]", "rate limiting")
	renderedLines, err := termRenderer.Render(in)
	if err != nil {
		return "", err
	}
	// For some reason, glamour adds escape sequences for the first data line of
	// each table. Remove these.
	renderedLines = unwantedEscapeSequenceRx.ReplaceAllString(renderedLines, "")
	renderedLines = trailingSpaceRx.ReplaceAllString(renderedLines, "\n")
	renderedLines = strings.Trim(renderedLines, "\n")
	return renderedLines, nil
}
