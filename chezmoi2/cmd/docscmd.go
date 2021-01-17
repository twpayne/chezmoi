// +build !nodocs

package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

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
		allDocsFilenames, err := docsFilenames()
		if err != nil {
			return err
		}
		var filenames []string
		for _, fn := range allDocsFilenames {
			if re.FindStringIndex(strings.ToLower(fn)) != nil {
				filenames = append(filenames, fn)
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

	documentData, err := doc(filename)
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

	return c.writeOutput(renderedData)
}
