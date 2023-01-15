package cmd

import "github.com/spf13/cobra"

func (c *Config) newCatConfigCmd() *cobra.Command {
	catConfigCmd := &cobra.Command{
		Use:     "cat-config",
		Short:   "Print the configuration file",
		Long:    mustLongHelp("cat-config"),
		Example: example("cat-config"),
		Args:    cobra.NoArgs,
		RunE:    c.runCatConfigCmd,
		Annotations: newAnnotations(
			requiresConfigDirectory,
			runsWithInvalidConfig,
		),
	}

	return catConfigCmd
}

func (c *Config) runCatConfigCmd(cmd *cobra.Command, args []string) error {
	data, err := c.baseSystem.ReadFile(c.configFileAbsPath)
	if err != nil {
		return err
	}
	return c.writeOutput(data)
}
