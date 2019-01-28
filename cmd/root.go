package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg"
)

var (
	config = Config{
		Umask: permValue(getUmask()),
		SourceVCS: sourceVCSConfig{
			Command: "git",
		},
	}
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCommand = &cobra.Command{
	Use:               "chezmoi",
	Short:             "Manage your dotfiles across multiple machines, securely",
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

	persistentFlags.StringVarP(&config.configFile, "config", "c", getDefaultConfigFile(x, homeDir), "config file")

	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", false, "dry run")
	viper.BindPFlag("dry-run", persistentFlags.Lookup("dry-run"))

	persistentFlags.StringVarP(&config.SourceDir, "source", "S", getDefaultSourceDir(x, homeDir), "source directory")
	viper.BindPFlag("source", persistentFlags.Lookup("source"))

	persistentFlags.StringVarP(&config.DestDir, "destination", "D", homeDir, "destination directory")
	viper.BindPFlag("destination", persistentFlags.Lookup("destination"))

	persistentFlags.VarP(&config.Umask, "umask", "u", "umask")
	viper.BindPFlag("umask", persistentFlags.Lookup("umask"))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	viper.BindPFlag("verbose", persistentFlags.Lookup("verbose"))

	cobra.OnInitialize(func() {
		viper.SetConfigFile(config.configFile)
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

func (c *Config) persistentPreRunRootE(fs vfs.FS, args []string) error {
	info, err := fs.Stat(c.SourceDir)
	switch {
	case err == nil && !info.IsDir():
		return fmt.Errorf("%s: not a directory", c.SourceDir)
	case err == nil && info.Mode().Perm() != 0700:
		fmt.Printf("%s: want permissions 0700, got 0%o\n", c.SourceDir, info.Mode().Perm())
	case os.IsNotExist(err):
	default:
		return err
	}
	return nil
}
