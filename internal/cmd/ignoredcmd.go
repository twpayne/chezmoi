package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

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

	return ignoredCmd
}

func (c *Config) runIgnoredCmd(
	cmd *cobra.Command,
	args []string,
	sourceState *chezmoi.SourceState,
) error {
	builder := strings.Builder{}
	for _, relPath := range sourceState.Ignored() {
		if _, err := builder.WriteString(relPath.String()); err != nil {
			return err
		}
		if err := builder.WriteByte('\n'); err != nil {
			return err
		}
	}
	return c.writeOutputString(builder.String())
}
