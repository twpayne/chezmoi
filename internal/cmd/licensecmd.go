package cmd

import (
	"bytes"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
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
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
		),
	}

	return licenseCmd
}

func (c *Config) runLicenseCmd(cmd *cobra.Command, args []string) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(styles.ASCIIStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	licenseMarkdown := bytes.TrimPrefix(docs.License, []byte("# License\n\n"))
	license, err := renderer.RenderBytes(licenseMarkdown)
	if err != nil {
		return err
	}

	return c.writeOutput(license)
}
