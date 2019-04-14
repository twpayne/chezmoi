package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

type dumpCmdConfig struct {
	format    string
	recursive bool
}

var dumpCmd = &cobra.Command{
	Use:     "dump [targets...]",
	Short:   "Write a dump of the target state to stdout",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runDumpCmd),
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	persistentFlags := dumpCmd.PersistentFlags()
	persistentFlags.StringVarP(&config.dump.format, "format", "f", "json", "format (JSON, TOML, or YAML)")
	persistentFlags.BoolVarP(&config.dump.recursive, "recursive", "r", true, "recursive")
}

func (c *Config) runDumpCmd(fs vfs.FS, args []string) error {
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
		entries, err := c.getEntries(fs, ts, args)
		if err != nil {
			return err
		}
		var concreteValues []interface{}
		for _, entry := range entries {
			entryConcreteValue, err := entry.ConcreteValue(ts.DestDir, ts.TargetIgnore.Match, ts.SourceDir, os.FileMode(c.Umask), c.dump.recursive)
			if err != nil {
				return err
			}
			if entryConcreteValue != nil {
				concreteValues = append(concreteValues, entryConcreteValue)
			}
		}
		concreteValue = concreteValues
	}
	return format(c.Stdout(), concreteValue)
}
