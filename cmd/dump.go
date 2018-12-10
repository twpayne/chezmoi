package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

type dumpCommandConfig struct {
	format string
}

var dumpCommand = &cobra.Command{
	Use:   "dump",
	Short: "Write a dump of the target state to stdout",
	RunE:  makeRunE(config.runDumpCommand),
}

func init() {
	rootCommand.AddCommand(dumpCommand)

	persistentFlags := dumpCommand.PersistentFlags()
	persistentFlags.StringVarP(&config.dump.format, "format", "f", "json", "format (JSON)")
}

func (c *Config) runDumpCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	var concreteValue interface{}
	if len(args) == 0 {
		concreteValue, err = targetState.ConcreteValue()
		if err != nil {
			return err
		}
	} else {
		entries, err := c.getEntries(targetState, args)
		if err != nil {
			return err
		}
		var concreteValues []interface{}
		for _, entry := range entries {
			entryConcreteValue, err := entry.ConcreteValue()
			if err != nil {
				return err
			}
			concreteValues = append(concreteValues, entryConcreteValue)
		}
		concreteValue = concreteValues
	}
	switch strings.ToLower(c.dump.format) {
	case "json":
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		return e.Encode(concreteValue)
	default:
		return fmt.Errorf("unknown format: %s", c.dump.format)
	}
}
