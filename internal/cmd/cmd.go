// Package cmd contains chezmoi's commands.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/multierr"

	"github.com/twpayne/chezmoi/v2/assets/chezmoi.io/docs/reference/commands"
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
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
	requiresWorkingTree          = "chezmoi_requires_working_tree"
	runsCommands                 = "chezmoi_runs_commands"
)

// Persistent state modes.
const (
	persistentStateModeEmpty         = "empty"
	persistentStateModeReadOnly      = "read-only"
	persistentStateModeReadMockWrite = "read-mock-write"
	persistentStateModeReadWrite     = "read-write"
)

var (
	noArgs = []string(nil)

	trailingSpaceRx = regexp.MustCompile(` +\n`)

	helps = make(map[string]*help)
)

// A VersionInfo contains a version.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

type help struct {
	longHelp string
	example  string
}

func init() {
	dirEntries, err := commands.FS.ReadDir(".")
	if err != nil {
		panic(err)
	}

	longHelpStyleConfig := glamour.ASCIIStyleConfig
	longHelpStyleConfig.Code.StylePrimitive.BlockPrefix = ""
	longHelpStyleConfig.Code.StylePrimitive.BlockSuffix = ""
	longHelpStyleConfig.Emph.BlockPrefix = ""
	longHelpStyleConfig.Emph.BlockSuffix = ""
	longHelpStyleConfig.H2.Prefix = ""
	longHelpTermRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(longHelpStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		panic(err)
	}

	exampleStyleConfig := glamour.ASCIIStyleConfig
	exampleStyleConfig.Code.StylePrimitive.BlockPrefix = ""
	exampleStyleConfig.Code.StylePrimitive.BlockSuffix = ""
	exampleStyleConfig.Document.Margin = nil
	exampleTermRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(exampleStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		panic(err)
	}

	for _, dirEntry := range dirEntries {
		command := strings.TrimSuffix(dirEntry.Name(), ".md")
		data, err := commands.FS.ReadFile(dirEntry.Name())
		if err != nil {
			panic(err)
		}
		help, err := extractHelp(command, data, longHelpTermRenderer, exampleTermRenderer)
		if err != nil {
			panic(err)
		}
		helps[command] = help
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (v VersionInfo) MarshalZerologObject(e *zerolog.Event) {
	e.Str("version", v.Version)
	e.Str("commit", v.Commit)
	e.Str("date", v.Date)
	e.Str("builtBy", v.BuiltBy)
}

// Main runs chezmoi and returns an exit code.
func Main(versionInfo VersionInfo, args []string) int {
	if err := runMain(versionInfo, args); err != nil {
		if s := err.Error(); s != "" {
			fmt.Fprintf(os.Stderr, "chezmoi: %s\n", s)
		}
		errExitCode := chezmoi.ExitCodeError(1)
		_ = errors.As(err, &errExitCode)
		return int(errExitCode)
	}
	return 0
}

// boolAnnotation returns whether cmd is annotated with key.
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

// example returns command's example.
func example(command string) string {
	help, ok := helps[command]
	if !ok {
		return ""
	}
	return help.example
}

// extractHelps returns the helps parse from r.
func extractHelp(
	command string, data []byte, longHelpTermRenderer, exampleTermRenderer *glamour.TermRenderer,
) (*help, error) {
	type stateType int
	const (
		stateReadTitle stateType = iota
		stateInLongHelp
		stateInOptions
		stateInExample
		stateInAdmonition
	)

	state := stateReadTitle
	var longHelpLines []string
	var exampleLines []string
	for _, line := range strings.Split(string(data), "\n") {
		switch state {
		case stateReadTitle:
			titleRx, err := regexp.Compile("# `" + command + "`")
			if err != nil {
				return nil, err
			}
			if titleRx.MatchString(line) {
				state = stateInLongHelp
			}
		case stateInLongHelp:
			switch {
			case strings.HasPrefix(line, "## "):
				state = stateInOptions
			case line == "!!! example":
				state = stateInExample
			case strings.HasPrefix(line, "!!!"):
				state = stateInAdmonition
			default:
				longHelpLines = append(longHelpLines, line)
			}
		case stateInOptions:
			if line == "!!! example" {
				state = stateInExample
			}
		case stateInExample:
			exampleLines = append(exampleLines, strings.TrimPrefix(line, "    "))
		case stateInAdmonition:
			if line == "!!! example" {
				state = stateInExample
			}
		}
	}

	longHelp, err := renderLines(longHelpLines, longHelpTermRenderer)
	if err != nil {
		return nil, err
	}
	example, err := renderLines(exampleLines, exampleTermRenderer)
	if err != nil {
		return nil, err
	}
	return &help{
		longHelp: "Description:\n" + longHelp,
		example:  example,
	}, nil
}

// renderLines renders lines, trimming extraneous whitespace.
func renderLines(lines []string, termRenderer *glamour.TermRenderer) (string, error) {
	renderedLines, err := termRenderer.Render(strings.Join(lines, "\n"))
	if err != nil {
		return "", err
	}
	renderedLines = trailingSpaceRx.ReplaceAllString(renderedLines, "\n")
	renderedLines = strings.Trim(renderedLines, "\n")
	return renderedLines, nil
}

// markPersistentFlagsRequired marks all of flags as required for cmd.
func markPersistentFlagsRequired(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		if err := cmd.MarkPersistentFlagRequired(flag); err != nil {
			panic(err)
		}
	}
}

// mustLongHelp returns the long help for command or panics if no long help
// exists.
func mustLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(fmt.Sprintf("missing long help for command %s", command))
	}
	return help.longHelp
}

// runMain runs chezmoi's main function.
func runMain(versionInfo VersionInfo, args []string) (err error) {
	var config *Config
	if config, err = newConfig(
		withVersionInfo(versionInfo),
	); err != nil {
		return err
	}
	defer func() {
		err = multierr.Append(err, config.close())
	}()
	err = config.execute(args)
	if errors.Is(err, bbolt.ErrTimeout) {
		// Translate bbolt timeout errors into a friendlier message. As the
		// persistent state is opened lazily, this error could occur at any
		// time, so it's easiest to intercept it here.
		err = errors.New("timeout obtaining persistent state lock, is another instance of chezmoi running?")
	}
	return
}
