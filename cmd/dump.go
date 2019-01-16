package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
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

func (c *Config) runDumpCommand(fs vfs.FS, args []string) error {
	format, ok := formatMap[strings.ToLower(c.dump.format)]
	if !ok {
		return fmt.Errorf("%s: unknown format", c.dump.format)
	}
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	var concreteValue interface{}
	if len(args) == 0 {
		concreteValue, err = ts.ConcreteValue(c.dump.recursive)
		if err != nil {
			return err
		}
	} else {
		entries, err := c.getEntries(ts, args)
		if err != nil {
			return err
		}
		var concreteValues []interface{}
		for _, entry := range entries {
			entryConcreteValue, err := entry.ConcreteValue(ts.DestDir, ts.TargetIgnore.Match, ts.SourceDir, c.dump.recursive)
			if err != nil {
				return err
			}
			if concreteValue != nil {
				concreteValues = append(concreteValues, entryConcreteValue)
			}
		}
		concreteValue = concreteValues
	}
	return format(os.Stdout, concreteValue)
}
