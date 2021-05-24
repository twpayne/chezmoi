package chezmoicmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/twpayne/chezmoi/v2/docs"
)

type docsCmdConfig struct {
	MaxWidth int    `mapstructure:"maxWidth"`
	Pager    string `mapstructure:"pager"`
}

func (c *Config) newDocsCmd() *cobra.Command {
	docsCmd := &cobra.Command{
		Use:     "docs [regexp]",
		Short:   "Print documentation",
		Long:    mustLongHelp("docs"),
		Example: example("docs"),
		Args:    cobra.MaximumNArgs(1),
		RunE:    c.runDocsCmd,
		Annotations: map[string]string{
			doesNotRequireValidConfig: "true",
		},
	}

	flags := docsCmd.Flags()
	flags.IntVar(&c.Docs.MaxWidth, "max-width", c.Docs.MaxWidth, "maximum output width")
	flags.StringVar(&c.Docs.Pager, "pager", c.Docs.Pager, "pager")

	return docsCmd
}

func (c *Config) runDocsCmd(cmd *cobra.Command, args []string) error {
	filename := "REFERENCE.md"
	if len(args) > 0 {
		pattern := args[0]
		re, err := regexp.Compile(strings.ToLower(pattern))
		if err != nil {
			return err
		}
		dirEntries, err := docs.FS.ReadDir(".")
		if err != nil {
			return err
		}
		var filenames []string
		for _, dirEntry := range dirEntries {
			fileInfo, err := dirEntry.Info()
			if err != nil {
				return err
			}
			if fileInfo.Mode().Type() != 0 {
				continue
			}
			if filename := dirEntry.Name(); re.FindStringIndex(strings.ToLower(filename)) != nil {
				filenames = append(filenames, filename)
			}
		}
		switch {
		case len(filenames) == 0:
			return fmt.Errorf("%s: no matching files", pattern)
		case len(filenames) == 1:
			filename = filenames[0]
		default:
			return fmt.Errorf("%s: ambiguous pattern, matches %s", pattern, strings.Join(filenames, ", "))
		}
	}

	file, err := docs.FS.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	documentData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	width := 80
	if stdout, ok := c.stdout.(*os.File); ok && term.IsTerminal(int(stdout.Fd())) {
		width, _, err = term.GetSize(int(stdout.Fd()))
		if err != nil {
			return err
		}
	}
	if c.Docs.MaxWidth != 0 && width > c.Docs.MaxWidth {
		width = c.Docs.MaxWidth
	}

	tr, err := glamour.NewTermRenderer(
		glamour.WithStyles(glamour.ASCIIStyleConfig),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return err
	}

	renderedData, err := tr.RenderBytes(documentData)
	if err != nil {
		return err
	}

	return c.pageOutputString(string(renderedData), c.Docs.Pager)
}
