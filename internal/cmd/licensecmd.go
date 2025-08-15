package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newLicenseCmd() *cobra.Command {
	licenseCmd := &cobra.Command{
		Use:               "license",
		Short:             "Print license",
		Long:              mustLongHelp("license"),
		Example:           example("license"),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              c.runLicenseCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}

	return licenseCmd
}

func (c *Config) runLicenseCmd(cmd *cobra.Command, args []string) error {
	return c.writeOutputString(license, 0o666)
}
