package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// Command annotations.
const (
	doesNotRequireValidConfig    = "chezmoi_does_not_require_valid_config"
	modifiesConfigFile           = "chezmoi_modifies_config_file"
	modifiesDestinationDirectory = "chezmoi_modifies_destination_directory"
	modifiesSourceDirectory      = "chezmoi_modifies_source_directory"
	persistentStateMode          = "chezmoi_persistent_state_mode"
	requiresConfigDirectory      = "chezmoi_requires_config_directory"
	requiresSourceDirectory      = "chezmoi_requires_source_directory"
	runsCommands                 = "chezmoi_runs_commands"
)

// Persistent state modes.
const (
	persistentStateModeEmpty         = "empty"
	persistentStateModeReadOnly      = "read-only"
	persistentStateModeReadMockWrite = "read-mock-write"
	persistentStateModeReadWrite     = "read-write"
)

var noArgs = []string(nil)

// An ErrExitCode indicates the the main program should exit with the given
// code.
type ErrExitCode int

func (e ErrExitCode) Error() string { return "" }

// A VersionInfo contains a version.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

// Main runs chezmoi and returns an exit code.
func Main(versionInfo VersionInfo, args []string) int {
	if err := runMain(versionInfo, args); err != nil {
		if s := err.Error(); s != "" {
			fmt.Fprintf(os.Stderr, "chezmoi: %s\n", s)
		}
		errExitCode := ErrExitCode(1)
		_ = errors.As(err, &errExitCode)
		return int(errExitCode)
	}
	return 0
}

func asset(name string) ([]byte, error) {
	asset, ok := assets[name]
	if !ok {
		return nil, fmt.Errorf("%s: not found", name)
	}
	return asset, nil
}

func boolAnnotation(cmd *cobra.Command, key string) bool {
	value, ok := cmd.Annotations[key]
	if !ok {
		return false
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}
	return boolValue
}

func example(command string) string {
	return helps[command].example
}

func mustLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(fmt.Sprintf("%s: no long help", command))
	}
	return help.long
}

func markPersistentFlagsRequired(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		if err := cmd.MarkPersistentFlagRequired(flag); err != nil {
			panic(err)
		}
	}
}

func runMain(versionInfo VersionInfo, args []string) error {
	config, err := newConfig(
		withVersionInfo(versionInfo),
	)
	if err != nil {
		return err
	}
	return config.execute(args)
}
