// +build !nodocs

package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:     "docs [pattern]",
	Args:    cobra.MaximumNArgs(1),
	Short:   "Print documentation",
	Long:    mustGetLongHelp("docs"),
	Example: getExample("docs"),
	RunE:    config.runDocsCmd,
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

func (c *Config) runDocsCmd(cmd *cobra.Command, args []string) error {
	filename := "REFERENCE.md"
	if len(args) > 0 {
		pattern := args[0]
		re, err := regexp.Compile(strings.ToLower(pattern))
		if err != nil {
			return err
		}
		docsFilenames, err := getDocsFilenames()
		if err != nil {
			return err
		}
		var filenames []string
		for _, fn := range docsFilenames {
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
	data, err := getDoc(filename)
	if err != nil {
		return err
	}
	_, err = c.Stdout().Write(data)
	return err
}
