// Package cmd contains chezmoi's commands.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"

	"github.com/twpayne/chezmoi/v2/assets/chezmoi.io/docs/reference/commands"
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
)

// GitHub owner and repo for chezmoi itself.
const (
	gitHubOwner = "twpayne"
	gitHubRepo  = "chezmoi"

	readSourceStateHookName = "read-source-state"
)

var (
	noArgs = []string(nil)

	deDuplicateErrorRx = regexp.MustCompile(`:\s+`)
	trailingSpaceRx    = regexp.MustCompile(` +\n`)

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
	err := runMain(versionInfo, args)
	if err != nil && strings.Contains(err.Error(), "unknown command") &&
		len(args) > 0 { // FIXME find a better way of detecting unknown commands
		if name, err2 := exec.LookPath("chezmoi-" + args[0]); err2 == nil {
			cmd := exec.Command(name, args[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
		}
	}
	if err != nil {
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
	seenComponents := make(map[string]struct{}, len(components))
	uniqueComponents := make([]string, 0, len(components))
	for _, component := range components {
		if _, ok := seenComponents[component]; ok {
			continue
		}
		uniqueComponents = append(uniqueComponents, component)
		seenComponents[component] = struct{}{}
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

// extractHelps returns the helps parse from r.
func extractHelp(command string, data []byte, longHelpTermRenderer, exampleTermRenderer *glamour.TermRenderer) (*help, error) {
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

// registerExcludeIncludeFlagCompletionFuncs registers the flag completion
// functions for the include and exclude flags of cmd. It panics on any error.
func registerExcludeIncludeFlagCompletionFuncs(cmd *cobra.Command) {
	if err := chezmoierrors.Combine(
		cmd.RegisterFlagCompletionFunc("exclude", chezmoi.EntryTypeSetFlagCompletionFunc),
		cmd.RegisterFlagCompletionFunc("include", chezmoi.EntryTypeSetFlagCompletionFunc),
	); err != nil {
		panic(err)
	}
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
		panic(command + ": missing long help")
	}
	return help.longHelp
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
	err = config.execute(args)
	if errors.Is(err, bbolt.ErrTimeout) {
		// Translate bbolt timeout errors into a friendlier message. As the
		// persistent state is opened lazily, this error could occur at any
		// time, so it's easiest to intercept it here.
		err = errors.New("timeout obtaining persistent state lock, is another instance of chezmoi running?")
	}
	return
}
