// Package cmd contains chezmoi's commands.
package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"

	"github.com/twpayne/chezmoi/v2/docs"
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

	commandsRx       = regexp.MustCompile(`^## Commands`)
	commandRx        = regexp.MustCompile("^### `(\\S+)`")
	exampleRx        = regexp.MustCompile("^#### `.+` examples")
	optionRx         = regexp.MustCompile("^#### `(-\\w|--\\w+)`")
	endOfCommandsRx  = regexp.MustCompile("^## ")
	horizontalRuleRx = regexp.MustCompile(`^---`)
	trailingSpaceRx  = regexp.MustCompile(` +\n`)

	helps map[string]*help
)

// An ExitCodeError indicates the the main program should exit with the given
// code.
type ExitCodeError int

func (e ExitCodeError) Error() string { return "" }

// A VersionInfo contains a version.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

type help struct {
	long    string
	example string
}

func init() {
	reference, err := docs.FS.ReadFile("REFERENCE.md")
	if err != nil {
		panic(err)
	}
	helps, err = extractHelps(bytes.NewReader(reference))
	if err != nil {
		panic(err)
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
		errExitCode := ExitCodeError(1)
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
func extractHelps(r io.Reader) (map[string]*help, error) {
	longStyleConfig := glamour.ASCIIStyleConfig
	longStyleConfig.Code.StylePrimitive.BlockPrefix = ""
	longStyleConfig.Code.StylePrimitive.BlockSuffix = ""
	longStyleConfig.Emph.BlockPrefix = ""
	longStyleConfig.Emph.BlockSuffix = ""
	longStyleConfig.H4.Prefix = ""
	longTermRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(longStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, err
	}

	examplesStyleConfig := glamour.ASCIIStyleConfig
	examplesStyleConfig.Document.Margin = nil
	examplesTermRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(examplesStyleConfig),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, err
	}

	type stateType int
	const (
		stateFindCommands stateType = iota
		stateFindFirstCommand
		stateInCommand
		stateFindExample
		stateInExample
	)

	var (
		state   = stateFindCommands
		builder = &strings.Builder{}
		h       *help
	)

	saveAndReset := func() error {
		var termRenderer *glamour.TermRenderer
		switch state {
		case stateInCommand, stateFindExample:
			termRenderer = longTermRenderer
		case stateInExample:
			termRenderer = examplesTermRenderer
		default:
			panic(fmt.Sprintf("%d: invalid state", state))
		}
		s, err := termRenderer.Render(builder.String())
		if err != nil {
			return err
		}
		s = trailingSpaceRx.ReplaceAllString(s, "\n")
		s = strings.Trim(s, "\n")
		switch state {
		case stateInCommand, stateFindExample:
			h.long = "Description:\n" + s
		case stateInExample:
			h.example = s
		default:
			panic(fmt.Sprintf("%d: invalid state", state))
		}
		builder.Reset()
		return nil
	}

	helps := make(map[string]*help)
	s := bufio.NewScanner(r)
FOR:
	for s.Scan() {
		switch state {
		case stateFindCommands:
			if commandsRx.MatchString(s.Text()) {
				state = stateFindFirstCommand
			}
		case stateFindFirstCommand:
			if m := commandRx.FindStringSubmatch(s.Text()); m != nil {
				h = &help{}
				helps[m[1]] = h
				state = stateInCommand
			}
		case stateInCommand, stateFindExample, stateInExample:
			switch m := commandRx.FindStringSubmatch(s.Text()); {
			case m != nil:
				if err := saveAndReset(); err != nil {
					return nil, err
				}
				h = &help{}
				helps[m[1]] = h
				state = stateInCommand
			case optionRx.MatchString(s.Text()):
				state = stateFindExample
			case exampleRx.MatchString(s.Text()):
				if err := saveAndReset(); err != nil {
					return nil, err
				}
				state = stateInExample
			case endOfCommandsRx.MatchString(s.Text()):
				if err := saveAndReset(); err != nil {
					return nil, err
				}
				break FOR
			case horizontalRuleRx.MatchString(s.Text()):
				if err := saveAndReset(); err != nil {
					return nil, err
				}
				state = stateFindFirstCommand
			case state != stateFindExample:
				if _, err := builder.WriteString(s.Text()); err != nil {
					return nil, err
				}
				if err := builder.WriteByte('\n'); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return helps, nil
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
	return help.long
}

// runMain runs chezmoi's main function.
func runMain(versionInfo VersionInfo, args []string) error {
	config, err := newConfig(
		withVersionInfo(versionInfo),
	)
	if err != nil {
		return err
	}
	defer config.close()
	switch err := config.execute(args); {
	case errors.Is(err, bbolt.ErrTimeout):
		// Translate bbolt timeout errors into a friendlier message. As the
		// persistent state is opened lazily, this error could occur at any
		// time, so it's easiest to intercept it here.
		return errors.New("timeout obtaining persistent state lock, is another instance of chezmoi running?")
	default:
		return err
	}
}
