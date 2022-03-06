package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/assets/chezmoi.io/docs"
)

func (c *Config) newLicenseCmd() *cobra.Command {
	licenseCmd := &cobra.Command{
		Use:     "license",
		Short:   "Print license",
		Long:    mustLongHelp("license"),
		Example: example("license"),
		Args:    cobra.NoArgs,
		RunE:    c.runLicenseCmd,
	}

	return licenseCmd
}

func (c *Config) runLicenseCmd(cmd *cobra.Command, args []string) error {
	return c.writeOutput(docs.License)
}
