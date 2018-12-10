package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
	yaml "gopkg.in/yaml.v2"
)

type dumpCommandConfig struct {
	format    string
	recursive bool
}

var dumpCommand = &cobra.Command{
	Use:   "dump",
	Short: "Write a dump of the target state to stdout",
	RunE:  makeRunE(config.runDumpCommand),
}

func init() {
	rootCommand.AddCommand(dumpCommand)

	persistentFlags := dumpCommand.PersistentFlags()
	persistentFlags.StringVarP(&config.dump.format, "format", "f", "json", "format (JSON or YAML)")
	persistentFlags.BoolVarP(&config.dump.recursive, "recursive", "r", true, "recursive")
}

func (c *Config) runDumpCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	var concreteValue interface{}
	if len(args) == 0 {
		concreteValue, err = targetState.ConcreteValue(c.dump.recursive)
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
			entryConcreteValue, err := entry.ConcreteValue(targetState.TargetDir, targetState.SourceDir, c.dump.recursive)
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
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(concreteValue)
	default:
		return fmt.Errorf("unknown format: %s", c.dump.format)
	}
}
