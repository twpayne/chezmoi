package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	vfs "github.com/twpayne/go-vfs"
	vfsafero "github.com/twpayne/go-vfsafero"
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
	if err := vfs.MkdirAll(c.mutator, filepath.Dir(c.configFile), 0777); err != nil {
		return err
	}

	if err := c.runEditor(c.configFile); err != nil {
		return err
	}

	// Warn the user of any errors reading the config file.
	v := viper.New()
	v.SetFs(vfsafero.NewAferoFS(c.fs))
	v.SetConfigFile(c.configFile)
	err := v.ReadInConfig()
	if err == nil {
		err = v.Unmarshal(&Config{})
	}
	if err != nil {
		cmd.Printf("warning: %s: %v\n", c.configFile, err)
	}

	return nil
}
