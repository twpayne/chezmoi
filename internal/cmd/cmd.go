// Package cmd contains chezmoi's commands.
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	bbolterrors "go.etcd.io/bbolt/errors"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

const readSourceStateHookName = "read-source-state"

var (
	noArgs = []string(nil)

	deDuplicateErrorRx = regexp.MustCompile(`:\s+`)
)

// A VersionInfo contains a version.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

func (v VersionInfo) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("version", v.Version),
		slog.String("commit", v.Commit),
		slog.String("date", v.Date),
		slog.String("builtBy", v.BuiltBy),
	)
}

// Main runs chezmoi and returns an exit code.
func Main(versionInfo VersionInfo, args []string) int {
	if err := runMain(versionInfo, args); err != nil {
		if errExitCode := chezmoi.ExitCodeError(0); errors.As(err, &errExitCode) {
			return int(errExitCode)
		}
		fmt.Fprintf(os.Stderr, "chezmoi: %s\n", deDuplicateError(err))
		return 1
	}
	return 0
}

// deDuplicateError returns err's human-readable string with duplicate components
// removed.
func deDuplicateError(err error) string {
	components := deDuplicateErrorRx.Split(err.Error(), -1)
	seenComponents := chezmoiset.NewWithCapacity[string](len(components))
	uniqueComponents := make([]string, 0, len(components))
	for _, component := range components {
		if seenComponents.Contains(component) {
			continue
		}
		uniqueComponents = append(uniqueComponents, component)
		seenComponents.Add(component)
	}
	return strings.Join(uniqueComponents, ": ")
}

// example returns command's example.
func example(command string) string {
	help, ok := helps[command]
	if !ok {
		return ""
	}
	return help.example
}

// markFlagsRequired marks all of flags as required for cmd.
func markFlagsRequired(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		must(cmd.MarkFlagRequired(flag))
	}
}

// must panics if err is not nil.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// mustValue panics if err is not nil, otherwise it returns value.
func mustValue[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// mustValues panics if err is not nil, otherwise it returns value1 and value2.
func mustValues[T1, T2 any](value1 T1, value2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return value1, value2
}

// mustLongHelp returns the long help for command or panics if no long help
// exists.
func mustLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(command + ": missing long help")
	}
	return help.longHelp
}

func ensureAllFlagsDocumented(cmd *cobra.Command, persistentFlags *pflag.FlagSet) {
	cmdName := cmd.Name()
	help, ok := helps[cmdName]
	if !ok && !cmd.Flags().HasFlags() {
		return
	}
	if !ok {
		panic(cmdName + ": missing flags")
	}
	// Check if all flags are documented.
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if _, ok := help.longFlags[flag.Name]; !ok {
			panic(fmt.Sprintf("%s: undocumented long flag --%s", cmdName, flag.Name))
		}
		if flag.Shorthand != "" {
			if _, ok := help.shortFlags[flag.Shorthand]; !ok {
				panic(fmt.Sprintf("%s: undocumented short flag -%s", cmdName, flag.Shorthand))
			}
		}
	})
	// Check if all documented flags exist.
	for flag := range help.longFlags {
		if cmd.Flags().Lookup(flag) == nil && persistentFlags.Lookup(flag) == nil {
			panic(fmt.Sprintf("%s: flag --%s documented but not implemented", cmdName, flag))
		}
	}
	for flag := range help.shortFlags {
		if cmd.Flags().ShorthandLookup(flag) == nil && persistentFlags.ShorthandLookup(flag) == nil {
			panic(fmt.Sprintf("%s: flag -%s documented but not implemented", cmdName, flag))
		}
	}
}

// runMain runs chezmoi's main function.
func runMain(versionInfo VersionInfo, args []string) (err error) {
	if versionInfo.Commit == "" || versionInfo.Date == "" {
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			var vcs, vcsRevision, vcsTime, vcsModified string
			for _, setting := range buildInfo.Settings {
				switch setting.Key {
				case "vcs":
					vcs = setting.Value
				case "vcs.revision":
					vcsRevision = setting.Value
				case "vcs.time":
					vcsTime = setting.Value
				case "vcs.modified":
					vcsModified = setting.Value
				}
			}
			if versionInfo.Commit == "" && vcs == "git" {
				versionInfo.Commit = vcsRevision
				if modified, err := strconv.ParseBool(vcsModified); err == nil && modified {
					versionInfo.Commit += "-dirty"
				}
			}
			if versionInfo.Date == "" {
				versionInfo.Date = vcsTime
			}
		}
	}

	var config *Config
	if config, err = newConfig(
		withVersionInfo(versionInfo),
	); err != nil {
		return err
	}
	defer chezmoierrors.CombineFunc(&err, config.Close)

	switch err = config.execute(args); {
	case errors.Is(err, bbolterrors.ErrTimeout):
		// Translate bbolt timeout errors into a friendlier message. As the
		// persistent state is opened lazily, this error could occur at any
		// time, so it's easiest to intercept it here.
		return errors.New("timeout obtaining persistent state lock, is another instance of chezmoi running?")
	case err != nil && strings.Contains(err.Error(), "unknown command") && len(args) > 0:
		// If the command is unknown then look for a plugin.
		if name, lookPathErr := exec.LookPath("chezmoi-" + args[0]); lookPathErr == nil {
			// The following is a bit of a hack, as cobra does not have a way to
			// call a function if a command is not found. We need to run the
			// pre- and post- run commands to set up the environment, so we
			// create a fake cobra.Command that corresponds to the name of the
			// plugin.
			cmd := &cobra.Command{
				Use: args[0],
				Annotations: newAnnotations(
					doesNotRequireValidConfig,
					persistentStateModeEmpty,
					runsCommands,
				),
			}
			if err := config.persistentPreRunRootE(cmd, args[1:]); err != nil {
				return err
			}
			pluginCmd := exec.Command(name, args[1:]...)
			pluginCmd.Stdin = os.Stdin
			pluginCmd.Stdout = os.Stdout
			pluginCmd.Stderr = os.Stderr
			err = config.run("", name, args[1:])
			if persistentPostRunRootEErr := config.persistentPostRunRootE(cmd, args[1:]); persistentPostRunRootEErr != nil {
				err = chezmoierrors.Combine(err, persistentPostRunRootEErr)
			}
		}
	}
	return
}
