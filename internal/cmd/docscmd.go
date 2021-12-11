package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"golang.org/x/term"

	"github.com/twpayne/chezmoi/v2/docs"
)

type docsCmdConfig struct {
	MaxWidth int    `mapstructure:"maxWidth"`
	Pager    string `mapstructure:"pager"`
}

func (c *Config) newDocsCmd() *cobra.Command {
	docsCmd := &cobra.Command{
		Use:               "docs [regexp]",
		Short:             "Print documentation",
		Long:              mustLongHelp("docs"),
		ValidArgsFunction: c.docsCmdValidArgs,
		Example:           example("docs"),
		Args:              cobra.MaximumNArgs(1),
		RunE:              c.runDocsCmd,
		Annotations: map[string]string{
			doesNotRequireValidConfig: "true",
		},
	}

	flags := docsCmd.Flags()
	flags.IntVar(&c.Docs.MaxWidth, "max-width", c.Docs.MaxWidth, "Set maximum output width")
	flags.StringVar(&c.Docs.Pager, "pager", c.Docs.Pager, "Set pager")

	return docsCmd
}

// docsCmdValidArgs returns the completions for the docs command.
func (c *Config) docsCmdValidArgs(
	cmd *cobra.Command, args []string, toComplete string,
) ([]string, cobra.ShellCompDirective) {
	var completions []string
	if err := fs.WalkDir(docs.FS, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dirEntry.IsDir() {
			return nil
		}
		completion := strings.ToLower(path)
		if strings.HasPrefix(completion, toComplete) {
			completions = append(completions, completion)
		}
		return nil
	}); err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}
	sort.Strings(completions)
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func (c *Config) runDocsCmd(cmd *cobra.Command, args []string) (err error) {
	filename := "REFERENCE.md"
	if len(args) > 0 {
		pattern := args[0]
		var re *regexp.Regexp
		if re, err = regexp.Compile(strings.ToLower(pattern)); err != nil {
			return
		}
		var dirEntries []fs.DirEntry
		if dirEntries, err = docs.FS.ReadDir("."); err != nil {
			return
		}
		var filenames []string
		for _, dirEntry := range dirEntries {
			var fileInfo fs.FileInfo
			if fileInfo, err = dirEntry.Info(); err != nil {
				return
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
			err = fmt.Errorf("%s: no matching files", pattern)
			return
		case len(filenames) == 1:
			filename = filenames[0]
		default:
			err = fmt.Errorf("%s: ambiguous pattern, matches %s", pattern, strings.Join(filenames, ", "))
			return
		}
	}

	var file fs.File
	if file, err = docs.FS.Open(filename); err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, file.Close())
	}()
	var documentData []byte
	if documentData, err = io.ReadAll(file); err != nil {
		return
	}

	width := 80
	if stdout, ok := c.stdout.(*os.File); ok && term.IsTerminal(int(stdout.Fd())) {
		if width, _, err = term.GetSize(int(stdout.Fd())); err != nil {
			return
		}
	}
	if c.Docs.MaxWidth != 0 && width > c.Docs.MaxWidth {
		width = c.Docs.MaxWidth
	}

	var termRenderer *glamour.TermRenderer
	if termRenderer, err = glamour.NewTermRenderer(
		glamour.WithStyles(glamour.ASCIIStyleConfig),
		glamour.WithWordWrap(width),
	); err != nil {
		return
	}

	var renderedData []byte
	if renderedData, err = termRenderer.RenderBytes(documentData); err != nil {
		return err
	}

	err = c.pageOutputString(string(renderedData), c.Docs.Pager)
	return
}
