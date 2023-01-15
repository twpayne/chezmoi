package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newEditConfigCmd() *cobra.Command {
	editConfigCmd := &cobra.Command{
		Use:     "edit-config",
		Short:   "Edit the configuration file",
		Long:    mustLongHelp("edit-config"),
		Example: example("edit-config"),
		Args:    cobra.NoArgs,
		RunE:    c.runEditConfigCmd,
		Annotations: newAnnotations(
			modifiesConfigFile,
			requiresConfigDirectory,
			runsCommands,
			runsWithInvalidConfig,
		),
	}

	return editConfigCmd
}

func (c *Config) runEditConfigCmd(cmd *cobra.Command, args []string) error {
	return c.runEditor([]string{c.configFileAbsPath.String()})
}
