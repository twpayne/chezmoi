package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var config = newConfig()

// Version information.
var (
	VersionStr string
	Commit     string
	Date       string
	BuiltBy    string
	Version    *semver.Version
)

var rootCmd = &cobra.Command{
	Use:               "chezmoi",
	Short:             "Manage your dotfiles across multiple machines, securely",
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: config.persistentPreRunRootE,
}

var (
	errExitFailure = errors.New("")
	initErr        error
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		initErr = err
		return
	}

	config.bds, err = xdg.NewBaseDirectorySpecification()
	if err != nil {
		initErr = err
		return
	}

	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.StringVarP(&config.configFile, "config", "c", getDefaultConfigFile(config.bds), "config file")
	panicOnError(rootCmd.MarkPersistentFlagFilename("config"))

	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", false, "dry run")
	panicOnError(viper.BindPFlag("dry-run", persistentFlags.Lookup("dry-run")))

	persistentFlags.BoolVar(&config.Follow, "follow", false, "follow symlinks")
	panicOnError(viper.BindPFlag("follow", persistentFlags.Lookup("follow")))

	persistentFlags.BoolVar(&config.Remove, "remove", false, "remove targets")
	panicOnError(viper.BindPFlag("remove", persistentFlags.Lookup("remove")))

	persistentFlags.StringVarP(&config.SourceDir, "source", "S", getDefaultSourceDir(config.bds), "source directory")
	panicOnError(viper.BindPFlag("source", persistentFlags.Lookup("source")))
	panicOnError(rootCmd.MarkPersistentFlagDirname("source"))

	persistentFlags.StringVarP(&config.DestDir, "destination", "D", homeDir, "destination directory")
	panicOnError(viper.BindPFlag("destination", persistentFlags.Lookup("destination")))
	panicOnError(rootCmd.MarkPersistentFlagDirname("destination"))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	panicOnError(viper.BindPFlag("verbose", persistentFlags.Lookup("verbose")))

	persistentFlags.StringVar(&config.Color, "color", "auto", "colorize diffs")
	panicOnError(viper.BindPFlag("color", persistentFlags.Lookup("color")))

	persistentFlags.BoolVar(&config.Debug, "debug", false, "write debug logs")
	panicOnError(viper.BindPFlag("debug", persistentFlags.Lookup("debug")))

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
				rootCmd.Printf("warning: %s: %v\n", config.configFile, config.err)
			}
			if config.GPGRecipient != "" {
				rootCmd.Printf("" +
					"warning: your config file uses gpgRecipient which will be deprecated in v2\n" +
					"warning: to disable this warning, set gpg.recipient in your config file instead\n",
				)
			}
			if config.SourceVCS.Command != "" && !config.SourceVCS.NotGit && !strings.Contains(filepath.Base(config.SourceVCS.Command), "git") {
				rootCmd.Printf("" +
					"warning: it looks like you are using a version control system that is not git which will be deprecated in v2\n" +
					"warning: please report this at https://github.com/twpayne/chezmoi/issues/459\n" +
					"warning: to disable this warning, set sourceVCS.notGit = true in your config file\n",
				)
			}
		case os.IsNotExist(err):
		default:
			initErr = err
		}
	})
}

// Execute executes the root command.
func Execute() error {
	if initErr != nil {
		return initErr
	}

	var versionComponents []string
	if VersionStr != "" {
		var err error
		Version, err = semver.NewVersion(strings.TrimPrefix(VersionStr, "v"))
		if err != nil {
			return err
		}
		versionComponents = append(versionComponents, VersionStr)
	} else {
		versionComponents = append(versionComponents, "dev")
	}
	if Commit != "" {
		versionComponents = append(versionComponents, "commit "+Commit)
	}
	if Date != "" {
		versionComponents = append(versionComponents, "built at "+Date)
	}
	if BuiltBy != "" {
		versionComponents = append(versionComponents, "built by "+BuiltBy)
	}
	rootCmd.Version = strings.Join(versionComponents, ", ")

	return rootCmd.Execute()
}

func (c *Config) persistentPreRunRootE(cmd *cobra.Command, args []string) error {
	if colored, err := strconv.ParseBool(c.Color); err == nil {
		c.colored = colored
	} else {
		switch c.Color {
		case "on":
			c.colored = true
		case "off":
			c.colored = false
		case "auto":
			if _, ok := os.LookupEnv("NO_COLOR"); ok {
				c.colored = false
			} else if stdout, ok := c.Stdout.(*os.File); ok {
				c.colored = terminal.IsTerminal(int(stdout.Fd()))
			} else {
				c.colored = false
			}
		default:
			return fmt.Errorf("invalid --color value: %s", c.Color)
		}
	}

	if c.colored {
		if err := enableVirtualTerminalProcessingOnWindows(c.Stdout); err != nil {
			return err
		}
	}

	c.fs = vfs.OSFS
	c.mutator = chezmoi.NewFSMutator(config.fs)
	if c.DryRun {
		c.mutator = chezmoi.NullMutator{}
	}
	if c.Debug {
		c.mutator = chezmoi.NewDebugMutator(c.mutator)
	}
	if c.Verbose {
		c.mutator = chezmoi.NewVerboseMutator(c.Stdout, c.mutator, c.colored, c.maxDiffDataSize)
	}

	if runtime.GOOS == "linux" && c.bds.RuntimeDir != "" {
		// Snap sets the $XDG_RUNTIME_DIR environment variable to
		// /run/user/$uid/snap.$snap_name, but does not create this directory.
		// Consequently, any spawned processes that need $XDG_DATA_DIR will
		// fail. As a work-around, create the directory if it does not exist.
		// See
		// https://forum.snapcraft.io/t/wayland-dconf-and-xdg-runtime-dir/186/13.
		if err := os.MkdirAll(c.bds.RuntimeDir, 0o700); err != nil {
			return err
		}
	}

	return nil
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
		panicOnError(cmd.MarkZshCompPositionalArgumentFile(from + i))
	}
}

func mustGetLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(fmt.Sprintf("no long help for %s", command))
	}
	return help.long
}
