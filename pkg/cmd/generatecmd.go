package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/assets/templates"
)

func (c *Config) newGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:       "generate file",
		Short:     "Generate a file for use with chezmoi",
		Long:      mustLongHelp("generate"),
		Example:   example("generate"),
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"install.sh"},
		RunE:      c.runGenerateCmd,
		Annotations: newAnnotations(
			runsWithInvalidConfig,
		),
	}

	return generateCmd
}

func (c *Config) runGenerateCmd(cmd *cobra.Command, args []string) error {
	builder := strings.Builder{}
	builder.Grow(16384)
	switch args[0] {
	case "install.sh":
		data, err := templates.FS.ReadFile("install.sh")
		if err != nil {
			return err
		}
		if _, err := builder.Write(data); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unsupported file", args[0])
	}
	return c.writeOutputString(builder.String())
}
