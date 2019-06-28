package cmd

import (
	"fmt"
	"regexp"
	"strings"

	packr "github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var docsCmd = &cobra.Command{
	Use:     "docs [pattern]",
	Args:    cobra.MaximumNArgs(1),
	Short:   "Print documentation",
	Long:    mustGetLongHelp("docs"),
	Example: getExample("docs"),
	RunE:    makeRunE(config.runDocsCmd),
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

func (c *Config) runDocsCmd(fs vfs.FS, args []string) error {
	box := packr.New("docs", "../docs")
	filename := "REFERENCE.md"
	if len(args) > 0 {
		pattern := args[0]
		re, err := regexp.Compile(strings.ToLower(pattern))
		if err != nil {
			return err
		}
		var filenames []string
		for _, fn := range box.List() {
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
	data, err := box.Find(filename)
	if err != nil {
		return err
	}
	_, err = c.Stdout().Write(data)
	return err
}
