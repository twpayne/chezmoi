package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg"
)

var (
	configFile string
	config     = Config{
		SourceVCSCommand: "git",
	}
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCommand = &cobra.Command{
	Use:               "chezmoi",
	Short:             "chezmoi is a tool for securely managing your dotfiles securely across multiple machines",
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: makeRunE(config.persistentPreRunRootE),
}

func init() {
	rootCommand.Version = fmt.Sprintf("%s, commit %s, built at %s", version, commit, date)

	homeDir, err := homedir.Dir()
	if err != nil {
		printErrorAndExit(err)
	}

	x, err := xdg.NewXDG()
	if err != nil {
		printErrorAndExit(err)
	}

	persistentFlags := rootCommand.PersistentFlags()

	persistentFlags.StringVarP(&configFile, "config", "c", getDefaultConfigFile(x, homeDir), "config file")

	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", false, "dry run")
	viper.BindPFlag("dry-run", persistentFlags.Lookup("dry-run"))

	persistentFlags.StringVarP(&config.SourceDir, "source", "s", getDefaultSourceDir(x, homeDir), "source directory")
	viper.BindPFlag("source", persistentFlags.Lookup("source"))

	persistentFlags.StringVarP(&config.TargetDir, "target", "t", homeDir, "target directory")
	viper.BindPFlag("target", persistentFlags.Lookup("target"))

	// FIXME umask should be printed in octal in help
	persistentFlags.IntVarP(&config.Umask, "umask", "u", getUmask(), "umask")
	viper.BindPFlag("umask", persistentFlags.Lookup("umask"))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	viper.BindPFlag("verbose", persistentFlags.Lookup("verbose"))

	cobra.OnInitialize(func() {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); os.IsNotExist(err) {
			return
		} else if err != nil {
			printErrorAndExit(err)
		}
		if err := viper.Unmarshal(&config); err != nil {
			printErrorAndExit(err)
		}
	})
}

// Execute executes the root command.
func Execute() {
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

func getDefaultConfigFile(x *xdg.XDG, homeDir string) string {
	// Search XDG config directories first.
	for _, configDir := range x.ConfigDirs {
		for _, extension := range viper.SupportedExts {
			configFilePath := filepath.Join(configDir, "chezmoi", "chezmoi."+extension)
			if _, err := os.Stat(configFilePath); err == nil {
				return configFilePath
			}
		}
	}
	// Search for ~/.chezmoi.* for backwards compatibility.
	for _, extension := range viper.SupportedExts {
		configFilePath := filepath.Join(homeDir, ".chezmoi."+extension)
		if _, err := os.Stat(configFilePath); err == nil {
			return configFilePath
		}
	}
	// Fallback to XDG default.
	return filepath.Join(x.ConfigHome, "chezmoi", "chezmoi.yaml")
}

func getDefaultSourceDir(x *xdg.XDG, homeDir string) string {
	// Check for XDG data directories first.
	for _, dataDir := range x.DataDirs {
		sourceDir := filepath.Join(dataDir, "chezmoi")
		if _, err := os.Stat(sourceDir); err == nil {
			return sourceDir
		}
	}
	// Check for ~/.chezmoi for backwards compatibility.
	sourceDir := filepath.Join(homeDir, ".chezmoi")
	if _, err := os.Stat(sourceDir); err == nil {
		return sourceDir
	}
	// Fallback to XDG default.
	return filepath.Join(x.DataHome, "chezmoi")
}
