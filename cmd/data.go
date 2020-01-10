package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type dataCmdConfig struct {
	format string
}

var dataCmd = &cobra.Command{
	Use:     "data",
	Args:    cobra.NoArgs,
	Short:   "Write the template data to stdout",
	Long:    mustGetLongHelp("data"),
	Example: getExample("data"),
	PreRunE: config.ensureNoError,
	RunE:    config.runDataCmd,
}

func init() {
	rootCmd.AddCommand(dataCmd)

	persistentFlags := dataCmd.PersistentFlags()
	persistentFlags.StringVarP(&config.data.format, "format", "f", "json", "format (JSON, TOML, or YAML)")
}

func (c *Config) runDataCmd(cmd *cobra.Command, args []string) error {
	format, ok := formatMap[strings.ToLower(c.data.format)]
	if !ok {
		return fmt.Errorf("%s: unknown format", c.data.format)
	}
	data, err := c.getData()
	if err != nil {
		return err
	}
	return format(c.Stdout(), data)
}
