package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"
	"golang.org/x/crypto/ssh/terminal"
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
	maxDiffDataSize:   1 * 1024 * 1024, // 1MB
	templateFuncs:     sprig.HermeticTxtFuncMap(),
	scriptStateBucket: []byte("script"),
}

// Version information.
var (
	devVersionStr = "dev"
	unknownStr    = "unknown"
	VersionStr    = devVersionStr
	Commit        = unknownStr
	Date          = unknownStr
	Version       *semver.Version
)

var rootCmd = &cobra.Command{
	Use:               "chezmoi",
	Short:             "Manage your dotfiles across multiple machines, securely",
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: config.persistentPreRunRootE,
}

func init() {
	if VersionStr != devVersionStr {
		var err error
		Version, err = semver.NewVersion(strings.TrimPrefix(VersionStr, "v"))
		if err != nil {
			printErrorAndExit(err)
		}
	}

	versionComponents := []string{VersionStr}
	if Commit != unknownStr {
		versionComponents = append(versionComponents, "commit "+Commit)
	}
	if Date != unknownStr {
		versionComponents = append(versionComponents, "built at "+Date)
	}
	rootCmd.Version = strings.Join(versionComponents, ", ")

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

	persistentFlags.BoolVar(&config.Follow, "follow", false, "follow symlinks")
	_ = viper.BindPFlag("follow", persistentFlags.Lookup("follow"))

	persistentFlags.BoolVar(&config.Remove, "remove", false, "remove targets")
	_ = viper.BindPFlag("remove", persistentFlags.Lookup("remove"))

	persistentFlags.StringVarP(&config.SourceDir, "source", "S", getDefaultSourceDir(config.bds), "source directory")
	_ = viper.BindPFlag("source", persistentFlags.Lookup("source"))

	persistentFlags.StringVarP(&config.DestDir, "destination", "D", homeDir, "destination directory")
	_ = viper.BindPFlag("destination", persistentFlags.Lookup("destination"))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	_ = viper.BindPFlag("verbose", persistentFlags.Lookup("verbose"))

	persistentFlags.StringVar(&config.Color, "color", "auto", "colorize diffs")
	_ = viper.BindPFlag("color", persistentFlags.Lookup("color"))

	persistentFlags.BoolVar(&config.Debug, "debug", false, "write debug logs")
	_ = viper.BindPFlag("debug", persistentFlags.Lookup("debug"))

	cobra.OnInitialize(func() {
		_, err := os.Stat(config.configFile)
		switch {
		case err == nil:
			viper.SetConfigFile(config.configFile)
			config.err = viper.ReadInConfig()
			if config.err == nil {
				config.err = viper.Unmarshal(&config)
			}
			if config.err == nil {
				config.err = config.validateData()
			}
			if config.err != nil {
				config.warn(fmt.Sprintf("%s: %v", config.configFile, config.err))
			}
		case os.IsNotExist(err):
		default:
			printErrorAndExit(err)
		}
	})
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErrorAndExit(err)
	}
}

//nolint:interfacer
func (c *Config) persistentPreRunRootE(cmd *cobra.Command, args []string) error {
	switch c.Color {
	case "on":
		c.colored = true
	case "off":
		c.colored = false
	case "auto":
		if stdout, ok := c.Stdout().(*os.File); ok {
			c.colored = terminal.IsTerminal(int(stdout.Fd()))
		} else {
			c.colored = false
		}
	default:
		return fmt.Errorf("invalid --color value: %s", c.Color)
	}

	if c.colored {
		if err := enableVirtualTerminalProcessingOnWindows(c.Stdout()); err != nil {
			return err
		}
	}

	c.fs = vfs.OSFS
	c.mutator = chezmoi.NewFSMutator(config.fs)
	if config.DryRun {
		c.mutator = chezmoi.NullMutator{}
	}
	if config.Debug {
		c.mutator = chezmoi.NewDebugMutator(c.mutator)
	}
	if config.Verbose {
		c.mutator = chezmoi.NewVerboseMutator(c.Stdout(), c.mutator, c.colored, c.maxDiffDataSize)
	}

	info, err := c.fs.Stat(c.SourceDir)
	switch {
	case err == nil && !info.IsDir():
		return fmt.Errorf("%s: not a directory", c.SourceDir)
	case err == nil:
		private, err := chezmoi.IsPrivate(c.fs, c.SourceDir, true)
		if err != nil {
			return err
		}
		if !private {
			c.warn(fmt.Sprintf("%s: not private, but should be", c.SourceDir))
		}
	case !os.IsNotExist(err):
		return err
	}

	// Apply any fixes for snap, if needed.
	return c.snapFix()
}

func getExample(command string) string {
	return helps[command].example
}

func markRemainingZshCompPositionalArgumentsAsFiles(cmd *cobra.Command, from int) {
	// As far as I can tell, there is no way to mark all remaining positional
	// arguments as files. Marking the first eight positional arguments as files
	// should be enough for everybody.
	// FIXME mark all remaining positional arguments as files
	for i := 0; i < 8; i++ {
		if err := cmd.MarkZshCompPositionalArgumentFile(from + i); err != nil {
			panic(err)
		}
	}
}

func mustGetLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(fmt.Sprintf("no long help for %s", command))
	}
	return help.long
}
