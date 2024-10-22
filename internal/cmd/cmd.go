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

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.etcd.io/bbolt"

	"github.com/twpayne/chezmoi/v2/assets/chezmoi.io/docs/reference/commands"
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

const readSourceStateHookName = "read-source-state"

var (
	noArgs = []string(nil)

	deDuplicateErrorRx = regexp.MustCompile(`:\s+`)
	trailingSpaceRx    = regexp.MustCompile(` +\n`)
	helpFlagsRx        = regexp.MustCompile("^### (?:`-([a-zA-Z])`, )?`--([a-zA-Z-]+)`")

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
	longHelp   string
	example    string
	longFlags  chezmoiset.Set[string]
	shortFlags chezmoiset.Set[string]
}

func init() {
	dirEntries, err := commands.FS.ReadDir(".")
	if err != nil {
		panic(err)
	}

	longHelpStyleConfig := styles.ASCIIStyleConfig
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

	exampleStyleConfig := styles.ASCIIStyleConfig
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

// extractHelp returns the helps parse from r.
func extractHelp(command string, data []byte, longHelpTermRenderer, exampleTermRenderer *glamour.TermRenderer) (*help, error) {
	type stateType int
	const (
		stateReadTitle stateType = iota
		stateInLongHelp
		stateInOptions
		stateInExamples
		stateInNotes
		stateInAdmonition
		stateInUnknownSection
	)

	state := stateReadTitle
	var longHelpLines []string
	var exampleLines []string
	longFlags := chezmoiset.New[string]()
	shortFlags := chezmoiset.New[string]()

	stateChange := func(line string, state *stateType) bool {
		switch {
		case line == "## Flags" || line == "## Common flags":
			*state = stateInOptions
			return true
		case line == "## Examples":
			*state = stateInExamples
			return true
		case line == "## Notes":
			*state = stateInNotes
			return true
		case strings.HasPrefix(line, "## "):
			*state = stateInUnknownSection
			return true
		}
		return false
	}

	for _, line := range strings.Split(string(data), "\n") {
		switch state {
		case stateReadTitle:
			titleRx, err := regexp.Compile("# `" + command + "`")
			if err != nil {
				return nil, err
			}
			if titleRx.MatchString(line) {
				state = stateInLongHelp
			} else {
				return nil, fmt.Errorf("expected title for '%s'", command)
			}
		case stateInLongHelp:
			switch {
			case stateChange(line, &state):
				break
			case strings.HasPrefix(line, "!!! "):
				state = stateInAdmonition
			default:
				longHelpLines = append(longHelpLines, line)
			}
		case stateInExamples:
			if !stateChange(line, &state) {
				exampleLines = append(exampleLines, line)
			}
		case stateInOptions:
			if !stateChange(line, &state) {
				matches := helpFlagsRx.FindStringSubmatch(line)
				if matches != nil {
					if matches[1] != "" {
						shortFlags.Add(matches[1])
					}
					longFlags.Add(matches[2])
				}
			}
		default:
			stateChange(line, &state)
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
		longHelp:   "Description:\n" + longHelp,
		example:    example,
		longFlags:  longFlags,
		shortFlags: shortFlags,
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

// markFlagsRequired marks all of flags as required for cmd.
func markFlagsRequired(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		if err := cmd.MarkFlagRequired(flag); err != nil {
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
	case errors.Is(err, bbolt.ErrTimeout):
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
