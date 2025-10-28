package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newEditConfigCmd() *cobra.Command {
	editConfigCmd := &cobra.Command{
		GroupID:           groupIDAdvanced,
		Use:               "edit-config",
		Short:             "Edit the configuration file",
		Long:              mustLongHelp("edit-config"),
		Example:           example("edit-config"),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              c.runEditConfigCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			modifiesConfigFile,
			persistentStateModeReadOnly,
			requiresConfigDirectory,
			runsCommands,
		),
	}

	return editConfigCmd
}

func (c *Config) runEditConfigCmd(cmd *cobra.Command, args []string) error {
	configFileAbsPath, err := c.getConfigFileAbsPath()
	if err != nil {
		return err
	}
	return c.runEditor([]string{configFileAbsPath.String()})
}
