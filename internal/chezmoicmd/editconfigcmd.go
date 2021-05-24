package chezmoicmd

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
		Annotations: map[string]string{
			modifiesConfigFile:      "true",
			requiresConfigDirectory: "true",
			runsCommands:            "true",
		},
	}

	return editConfigCmd
}

func (c *Config) runEditConfigCmd(cmd *cobra.Command, args []string) error {
	return c.runEditor([]string{string(c.configFileAbsPath)})
}
