package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/coreos/go-semver/semver"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"
)

var config = Config{
	Umask: permValue(getUmask()),
	Color: "auto",
	SourceVCS: sourceVCSConfig{
		Command: "git",
	},
	Merge: mergeConfig{
		Command: "vimdiff",
	},
	templateFuncs: sprig.HermeticTxtFuncMap(),
}

// Version information.
var (
	devVersionStr = "dev"
	VersionStr    = devVersionStr
	Commit        = "unknown"
	Date          = "unknown"
	Version       *semver.Version
)

var rootCmd = &cobra.Command{
	Use:               "chezmoi",
	Short:             "Manage your dotfiles across multiple machines, securely",
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: makeRunE(config.persistentPreRunRootE),
}

func init() {
	if VersionStr != devVersionStr {
		var err error
		Version, err = semver.NewVersion(strings.TrimPrefix(VersionStr, "v"))
		if err != nil {
			printErrorAndExit(err)
		}
	}

	rootCmd.Version = fmt.Sprintf("%s, commit %s, built at %s", VersionStr, Commit, Date)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		printErrorAndExit(err)
	}

	config.bds, err = xdg.NewBaseDirectorySpecification()
	if err != nil {
		printErrorAndExit(err)
	}

	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.StringVarP(&config.configFile, "config", "c", getDefaultConfigFile(config.bds), "config file")

	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", false, "dry run")
	_ = viper.BindPFlag("dry-run", persistentFlags.Lookup("dry-run"))

	persistentFlags.StringVarP(&config.SourceDir, "source", "S", getDefaultSourceDir(config.bds), "source directory")
	_ = viper.BindPFlag("source", persistentFlags.Lookup("source"))

	persistentFlags.StringVarP(&config.DestDir, "destination", "D", homeDir, "destination directory")
	_ = viper.BindPFlag("destination", persistentFlags.Lookup("destination"))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	_ = viper.BindPFlag("verbose", persistentFlags.Lookup("verbose"))

	persistentFlags.StringVar(&config.Color, "color", "auto", "colorize diffs")
	_ = viper.BindPFlag("color", persistentFlags.Lookup("color"))

	cobra.OnInitialize(func() {
		if _, err := os.Stat(config.configFile); os.IsNotExist(err) {
			return
		}
		viper.SetConfigFile(config.configFile)
		config.err = viper.ReadInConfig()
		if config.err == nil {
			config.err = viper.Unmarshal(&config)
		}
		if config.err != nil {
			config.warn(fmt.Sprintf("%s: %v", config.configFile, config.err))
		}
		persistentStateFile := getPersistentStateFile(config.bds, config.configFile)
		persistentState, err := chezmoi.NewBoltPersistentState(vfs.OSFS, persistentStateFile)
		if err != nil {
			printErrorAndExit(err)
		}
		config.persistentState = persistentState
	})
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErrorAndExit(err)
	}
}

func (c *Config) persistentPreRunRootE(fs vfs.FS, args []string) error {
	switch c.Color {
	case "on":
		c.colored = true
	case "off":
		c.colored = false
	case "auto":
		if stdout, ok := c.Stdout().(*os.File); ok {
			c.colored = isatty.IsTerminal(stdout.Fd())
		} else {
			c.colored = false
		}
	default:
		return fmt.Errorf("invalid --color value: %s", c.Color)
	}

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

func getExample(command string) string {
	return helps[command].example
}

func mustGetLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(fmt.Sprintf("no long help for %s", command))
	}
	return help.long
}
