package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var managedCmd = &cobra.Command{
	Use:     "managed",
	Args:    cobra.NoArgs,
	Short:   "List the managed files in the destination directory",
	Long:    mustGetLongHelp("managed"),
	Example: getExample("managed"),
	PreRunE: config.ensureNoError,
	RunE:    config.runManagedCmd,
}

func init() {
	rootCmd.AddCommand(managedCmd)
}

func recurseEntries(in []chezmoi.Entry) []string {
	out := make([]string, 0, len(in))
	for _, entry := range in {
		switch v := entry.(type) {
		case *chezmoi.Dir:
			entries := make([]chezmoi.Entry, 0, len(v.Entries))
			for _, entry := range v.Entries {
				entries = append(entries, entry)
			}
			results := recurseEntries(entries)
			out = append(out, results...)
		case *chezmoi.File, *chezmoi.Symlink:
			out = append(out, v.TargetName())
		}
	}
	return out
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}
	entries := make([]chezmoi.Entry, 0, len(ts.Entries))
	for _, entry := range ts.Entries {
		entries = append(entries, entry)
	}
	allManaged := recurseEntries(entries)
	for _, tn := range allManaged {
		path := filepath.Join(ts.DestDir, tn)
		entry, err := ts.Get(c.fs, path)
		if err != nil {
			return err
		}
		managed := entry != nil
		ignored := ts.TargetIgnore.Match(tn)
		if !ignored && managed {
			fmt.Fprintln(c.Stdout, path)
		}
	}
	return nil
}
