package cmd

import (
	"encoding/json"
	"fmt"
	"io"
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

var dumpFuncFormatMap = map[string]func(io.Writer, interface{}) error{
	"json": func(w io.Writer, value interface{}) error {
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		return e.Encode(value)
	},
	"yaml": func(w io.Writer, value interface{}) error {
		return yaml.NewEncoder(w).Encode(value)
	},
}

func init() {
	rootCommand.AddCommand(dumpCommand)

	persistentFlags := dumpCommand.PersistentFlags()
	persistentFlags.StringVarP(&config.dump.format, "format", "f", "json", "format (JSON or YAML)")
	persistentFlags.BoolVarP(&config.dump.recursive, "recursive", "r", true, "recursive")
}

func (c *Config) runDumpCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	dumpFunc, ok := dumpFuncFormatMap[strings.ToLower(c.dump.format)]
	if !ok {
		return fmt.Errorf("unknown format: %s", c.dump.format)
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
			entryConcreteValue, err := entry.ConcreteValue(ts.TargetDir, ts.SourceDir, c.dump.recursive)
			if err != nil {
				return err
			}
			concreteValues = append(concreteValues, entryConcreteValue)
		}
		concreteValue = concreteValues
	}
	return dumpFunc(os.Stdout, concreteValue)
}
