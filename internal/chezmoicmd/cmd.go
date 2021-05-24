// Package cmd contains chezmoi's commands.
package chezmoicmd

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

	commandsRx      = regexp.MustCompile(`^## Commands`)
	commandRx       = regexp.MustCompile("^### `(\\S+)`")
	exampleRx       = regexp.MustCompile("^#### `.+` examples")
	optionRx        = regexp.MustCompile("^#### `(-\\w|--\\w+)`")
	endOfCommandsRx = regexp.MustCompile(`^## `)
	trailingSpaceRx = regexp.MustCompile(` +\n`)

	helps map[string]*help
)

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

	var (
		state = "find-commands"
		sb    = &strings.Builder{}
		h     *help
	)

	saveAndReset := func() error {
		var tr *glamour.TermRenderer
		switch state {
		case "in-command", "find-example":
			tr = longTermRenderer
		case "in-example":
			tr = examplesTermRenderer
		default:
			panic(fmt.Sprintf("%s: invalid state", state))
		}
		s, err := tr.Render(sb.String())
		if err != nil {
			return err
		}
		s = trailingSpaceRx.ReplaceAllString(s, "\n")
		s = strings.Trim(s, "\n")
		switch state {
		case "in-command", "find-example":
			h.long = "Description:\n" + s
		case "in-example":
			h.example = s
		default:
			panic(fmt.Sprintf("%s: invalid state", state))
		}
		sb.Reset()
		return nil
	}

	helps := make(map[string]*help)
	s := bufio.NewScanner(r)
FOR:
	for s.Scan() {
		switch state {
		case "find-commands":
			if commandsRx.MatchString(s.Text()) {
				state = "find-first-command"
			}
		case "find-first-command":
			if m := commandRx.FindStringSubmatch(s.Text()); m != nil {
				h = &help{}
				helps[m[1]] = h
				state = "in-command"
			}
		case "in-command", "find-example", "in-example":
			m := commandRx.FindStringSubmatch(s.Text())
			switch {
			case m != nil:
				if err := saveAndReset(); err != nil {
					return nil, err
				}
				h = &help{}
				helps[m[1]] = h
				state = "in-command"
			case optionRx.MatchString(s.Text()):
				state = "find-example"
			case exampleRx.MatchString(s.Text()):
				if err := saveAndReset(); err != nil {
					return nil, err
				}
				state = "in-example"
			case endOfCommandsRx.MatchString(s.Text()):
				if err := saveAndReset(); err != nil {
					return nil, err
				}
				break FOR
			case state != "find-example":
				if _, err := sb.WriteString(s.Text()); err != nil {
					return nil, err
				}
				if err := sb.WriteByte('\n'); err != nil {
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
