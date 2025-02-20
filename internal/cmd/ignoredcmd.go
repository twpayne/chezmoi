package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type ignoredCmdConfig struct {
	nulPathSeparator bool
	tree             bool
}

func (c *Config) newIgnoredCmd() *cobra.Command {
	ignoredCmd := &cobra.Command{
		Use:     "ignored",
		Short:   "Print ignored targets",
		Long:    mustLongHelp("ignored"),
		Example: example("ignored"),
		Args:    cobra.NoArgs,
		RunE:    c.makeRunEWithSourceState(c.runIgnoredCmd),
		Annotations: newAnnotations(
			persistentStateModeReadMockWrite,
		),
	}

	ignoredCmd.Flags().BoolVarP(&c.ignored.tree, "tree", "t", c.ignored.tree, "Print paths as a tree")
	ignoredCmd.Flags().
		BoolVarP(&c.ignored.nulPathSeparator, "nul-path-separator", "0", c.ignored.nulPathSeparator, "Use the NUL character as a path separator")

	return ignoredCmd
}

func (c *Config) runIgnoredCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	return c.writePaths(stringersToStrings(sourceState.Ignored()), writePathsOptions{
		nulPathSeparator: c.ignored.nulPathSeparator,
		tree:             c.ignored.tree,
	})
}
