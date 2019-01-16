package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

type dataCommandConfig struct {
	format string
}

var dataCommand = &cobra.Command{
	Use:   "data",
	Short: "Write the template data to stdout",
	RunE:  makeRunE(config.runDataCommand),
}

func init() {
	rootCommand.AddCommand(dataCommand)

	persistentFlags := dataCommand.PersistentFlags()
	persistentFlags.StringVarP(&config.data.format, "format", "f", "json", "format (JSON or YAML)")
}

func (c *Config) runDataCommand(fs vfs.FS, args []string) error {
	format, ok := formatMap[strings.ToLower(c.data.format)]
	if !ok {
		return fmt.Errorf("%s: unknown format", c.data.format)
	}
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	return format(os.Stdout, ts.Data)
}
