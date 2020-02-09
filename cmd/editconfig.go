package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/configparser"
	vfs "github.com/twpayne/go-vfs"
)

var editConfigCommand = &cobra.Command{
	Use:     "edit-config",
	Args:    cobra.NoArgs,
	Short:   "Edit the configuration file",
	Long:    mustGetLongHelp("edit-config"),
	Example: getExample("edit-config"),
	RunE:    config.runEditConfigCmd,
}

func init() {
	rootCmd.AddCommand(editConfigCommand)
}

func (c *Config) runEditConfigCmd(cmd *cobra.Command, args []string) error {
	configFileName := c.getConfigFileName()
	if err := vfs.MkdirAll(c.mutator, filepath.Dir(configFileName), 0777); err != nil {
		return err
	}

	if err := c.runEditor(configFileName); err != nil {
		return err
	}

	// Warn the user of any errors reading the config file.
	configFile, err := os.Open(configFileName)
	if err != nil {
		c.warn(fmt.Sprintf("%s: %v", configFileName, err))
		return nil
	}
	defer configFile.Close()
	if err := configparser.ParseConfig(configFile, &Config{}); err != nil {
		c.warn(fmt.Sprintf("%s: %v", configFileName, err))
	}

	return nil
}
