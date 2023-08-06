package cobracompletefig

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var includeHidden bool
var figGenCmdUse string = ""

var generateCommandArgs func(*cobra.Command) Args

type Opts struct {
	Use                 string
	Short               string
	Visible             bool
	Long                string
	commandArgGenerator func(*cobra.Command) Args
}

func CreateCompletionSpecCommand(options ...Opts) * cobra.Command {
	Use := "generate-fig-spec"
	Aliases := []string{"generateFigSpec", "genFigSpec"}
	Short := "Generate a fig spec"
	Hidden := true
	Long := `
Fig is a tool for your command line that adds autocomplete.
This command generates a TypeScript file with the skeleton
Fig autocomplete spec for your Cobra CLI.
`
	if len(options) > 0 {
		if options[0].Use != "" {
			Use = options[0].Use
		}
		if options[0].Short != "" {
			Short = options[0].Short
		}
		if options[0].Long != "" {
			Long = options[0].Long
		}
		if options[0].Visible {
			Hidden = false
		}
		if options[0].commandArgGenerator != nil {
			generateCommandArgs = options[0].commandArgGenerator
		}
	}
	figGenCmdUse = Use
	var cmd = &cobra.Command{
		Use:    Use,
		Aliases: Aliases,
		Short:  Short,
		Hidden: Hidden,
		Long:   Long,
		Run: func(cmd *cobra.Command, args []string) {
			root := cmd.Root()
			spec := GenerateCompletionSpec(root)
			fmt.Println(spec.ToTypescript())
		},
	}
	cmd.Flags().BoolVar(
		&includeHidden, "include-hidden", false,
		"Include hidden commands in generated Fig autocomplete spec")
	return cmd
}

func GenerateCompletionSpec(root *cobra.Command) Spec {
	opts := append(options(root.LocalNonPersistentFlags(), false), options(root.PersistentFlags(), true)...)
	opts = append(opts, makeHelpOption())
	spec := Spec{
		Subcommand: &Subcommand{
			BaseSuggestion: &BaseSuggestion{
				description: root.Short,
			},
			options:     opts,
			subcommands: append(subcommands(root, false, Options{}), makeHelpCommand(root)), // We assume CLI is using default help command
			args:        commandArguments(root),
		},
		name: root.Name(),
	}
	return spec
}

func subcommands(cmd *cobra.Command, overrideOptions bool, overrides Options) Subcommands {
	var subs []Subcommand
	for _, sub := range cmd.Commands() {
		if sub.Name() == "help" || (!includeHidden && sub.Hidden) || (sub.Use == figGenCmdUse) {
			continue
		}
		var opts Options
		if overrideOptions {
			opts = overrides
		} else {
			opts = append(options(sub.LocalNonPersistentFlags(), false), options(sub.PersistentFlags(), true)...)
		}
		subs = append(subs, Subcommand{
			BaseSuggestion: &BaseSuggestion{
				description: sub.Short,
				hidden: sub.Hidden,
			},
			name:        append(sub.Aliases, sub.Name()),
			options:     opts,
			subcommands: subcommands(sub, overrideOptions, overrides),
			args:        commandArguments(sub),
		})
	}
	return subs
}

func isFlagRepeatable(flag *pflag.Flag) bool {
	return strings.Contains(flag.Value.Type(), "Slice") || strings.Contains(flag.Value.Type(), "Array")
}

func options(flagSet *pflag.FlagSet, persistent bool) []Option {
	var opts []Option
	attachFlags := func(flag *pflag.Flag) {

		option := Option{
			BaseSuggestion: &BaseSuggestion{
				description: flag.Usage,
				hidden: flag.Hidden,
			},
			name:         []string{fmt.Sprintf("--%v", flag.Name)},
			isRepeatable: isFlagRepeatable(flag),
		}
		if flag.Shorthand != "" {
			option.name = append(option.name, fmt.Sprintf("-%v", flag.Shorthand))
		}
		if persistent != false {
			option.isPersistent = true
		}
		requiredAnnotation, found := flag.Annotations[cobra.BashCompOneRequiredFlag]
		if found && requiredAnnotation[0] == "true" {
			option.isRequired = true
		}
		option.args = flagArguments(flag)
		opts = append(opts, option)
	}

	flagSet.VisitAll(attachFlags)
	return opts
}

/*
 * In Cobra, you only specify the number of arguments.
 * Not sure how we want to handle this (if at all)
 * https://github.com/spf13/cobra/blob/v1.2.1/user_guide.md#positional-and-custom-arguments
 */
func commandArguments(cmd *cobra.Command) []Arg {
	if generateCommandArgs != nil {
		return generateCommandArgs(cmd)
	}
	return []Arg{}
}

func flagArguments(flag *pflag.Flag) []Arg {
	var args []Arg
	defaultVal := flag.DefValue
	if defaultVal == "[]" {
		defaultVal = ""
	}
	if flag.Value.Type() != "bool" {
		arg := Arg{
			name:       flag.Name,
			defaultVal: defaultVal,
		}
		_, foundFilenameAnnotation := flag.Annotations[cobra.BashCompFilenameExt]
		if foundFilenameAnnotation {
			arg.template = append(arg.template, FILEPATHS)
		}
		_, foundDirectoryAnnotation := flag.Annotations[cobra.BashCompSubdirsInDir]
		if foundDirectoryAnnotation {
			arg.template = append(arg.template, FOLDERS)
		}
		args = append(args, arg)
	}
	return args
}

func makeHelpCommand(root *cobra.Command) Subcommand {
	return Subcommand{
		BaseSuggestion: &BaseSuggestion{
			description: "Help about any command",
		},
		name:        []string{"help"},
		subcommands: subcommands(root, true, []Option{}),
	}
}

func makeHelpOption() Option {
	return Option{
		BaseSuggestion: &BaseSuggestion{
			description: fmt.Sprintf("Display help"),
		},
		isPersistent: true,
		name: []string{"--help", "-h"},
	}
}
