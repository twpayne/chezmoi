package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

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

type managedCmdConfig struct {
	include []string
}

func init() {
	rootCmd.AddCommand(managedCmd)

	persistentFlags := managedCmd.PersistentFlags()
	persistentFlags.StringSliceVarP(&config.managed.include, "include", "i", []string{"dirs", "files", "symlinks"}, "include")
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}

	var (
		includeDirs     = false
		includeFiles    = false
		includeSymlinks = false
	)
	for _, what := range c.managed.include {
		switch what {
		case "dirs", "d":
			includeDirs = true
		case "files", "f":
			includeFiles = true
		case "symlinks", "s":
			includeSymlinks = true
		default:
			return fmt.Errorf("unrecognized include: %q", what)
		}
	}

	allEntries := ts.AllEntries()

	targetNames := make([]string, 0, len(allEntries))
	for _, entry := range allEntries {
		if _, ok := entry.(*chezmoi.Dir); ok && !includeDirs {
			continue
		}
		if _, ok := entry.(*chezmoi.File); ok && !includeFiles {
			continue
		}
		if _, ok := entry.(*chezmoi.Symlink); ok && !includeSymlinks {
			continue
		}
		targetNames = append(targetNames, entry.TargetName())
	}

	sort.Strings(targetNames)
	for _, targetName := range targetNames {
		if ts.TargetIgnore.Match(targetName) {
			continue
		}
		fmt.Fprintln(c.Stdout, filepath.Join(ts.DestDir, targetName))
	}

	return nil
}
