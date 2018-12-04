package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/go-vfs"
)

var (
	configFile string
	config     Config
)

var rootCommand = &cobra.Command{
	Use:               "chezmoi",
	Short:             "chezmoi is a tool for managing your home directory across multiple machines",
	PersistentPreRunE: makeRunE(config.persistentPreRunRootE),
}

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		printErrorAndExit(err)
	}

	persistentFlags := rootCommand.PersistentFlags()

	persistentFlags.StringVarP(&configFile, "config", "c", filepath.Join(homeDir, ".chezmoi"), "config file")

	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", false, "dry run")
	viper.BindPFlag("dry-run", persistentFlags.Lookup("dry-run"))

	persistentFlags.StringVarP(&config.SourceDir, "source", "s", filepath.Join(homeDir, ".chezmoi"), "source directory")
	viper.BindPFlag("source", persistentFlags.Lookup("source"))

	persistentFlags.StringVarP(&config.TargetDir, "target", "t", homeDir, "target directory")
	viper.BindPFlag("target", persistentFlags.Lookup("target"))

	// FIXME umask should be printed in octal in help
	persistentFlags.IntVarP(&config.Umask, "umask", "u", getUmask(), "umask")
	viper.BindPFlag("umask", persistentFlags.Lookup("umask"))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	viper.BindPFlag("verbose", persistentFlags.Lookup("verbose"))

	cobra.OnInitialize(func() {
		// FIXME once https://github.com/spf13/viper/pull/601 is merged, we can
		// use viper.SetConfigName instead of looping over possible config file
		// names ourself.
		for _, extension := range append([]string{""}, viper.SupportedExts...) {
			configFileName := configFile
			if extension != "" {
				configFileName += "." + extension
			}
			if info, err := os.Stat(configFileName); err != nil || !info.Mode().IsRegular() {
				continue
			}
			viper.SetConfigFile(configFileName)
			if err := viper.ReadInConfig(); err != nil {
				printErrorAndExit(err)
			}
			if err := viper.Unmarshal(&config); err != nil {
				printErrorAndExit(err)
			}
			return
		}
	})
}

// Execute executes the root command.
func Execute(version Version) {
	config.version = version
	if err := rootCommand.Execute(); err != nil {
		printErrorAndExit(err)
	}
}

func (c *Config) persistentPreRunRootE(fs vfs.FS, command *cobra.Command, args []string) error {
	info, err := fs.Stat(c.SourceDir)
	switch {
	case err == nil && !info.Mode().IsDir():
		return fmt.Errorf("%s: not a directory", c.SourceDir)
	case err == nil && info.Mode()&os.ModePerm != 0700:
		fmt.Printf("%s: want permissions 0700, got 0%o\n", c.SourceDir, info.Mode()&os.ModePerm)
	case os.IsNotExist(err):
	default:
		return err
	}
	return nil
}
