package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type ignoredCmdConfig struct {
	tree bool
}

func (c *Config) newIgnoredCmd() *cobra.Command {
	ignoredCmd := &cobra.Command{
		Use:         "ignored",
		Short:       "Print ignored targets",
		Long:        mustLongHelp("ignored"),
		Example:     example("ignored"),
		Args:        cobra.NoArgs,
		RunE:        c.makeRunEWithSourceState(c.runIgnoredCmd),
		Annotations: newAnnotations(),
	}

	ignoredCmd.Flags().BoolVarP(&c.ignored.tree, "tree", "t", c.ignored.tree, "Print paths as a tree")

	return ignoredCmd
}

func (c *Config) runIgnoredCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	relPaths := sourceState.Ignored()
	paths := make([]string, 0, len(relPaths))
	for _, relPath := range relPaths {
		paths = append(paths, relPath.String())
	}
	return c.writePaths(paths, writePathsOptions{
		tree: c.ignored.tree,
	})
}
