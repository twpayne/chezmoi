package cmd

import (
	"fmt"
	"go/build"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type executeTemplateCmdConfig struct {
	init         bool
	promptBool   map[string]string
	promptInt    map[string]int
	promptString map[string]string
	stdinIsATTY  bool
}

func (c *Config) newExecuteTemplateCmd() *cobra.Command {
	executeTemplateCmd := &cobra.Command{
		Use:     "execute-template [template]...",
		Short:   "Execute the given template(s)",
		Long:    mustLongHelp("execute-template"),
		Example: example("execute-template"),
		RunE:    c.runExecuteTemplateCmd,
	}

	flags := executeTemplateCmd.Flags()
	flags.BoolVarP(&c.executeTemplate.init, "init", "i", c.executeTemplate.init, "Simulate chezmoi init")
	flags.StringToStringVar(&c.executeTemplate.promptBool, "promptBool", c.executeTemplate.promptBool, "Simulate promptBool") //nolint:lll
	flags.StringToIntVar(&c.executeTemplate.promptInt, "promptInt", c.executeTemplate.promptInt, "Simulate promptInt")
	flags.StringToStringVarP(&c.executeTemplate.promptString, "promptString", "p", c.executeTemplate.promptString, "Simulate promptString") //nolint:lll
	flags.BoolVar(&c.executeTemplate.stdinIsATTY, "stdinisatty", c.executeTemplate.stdinIsATTY, "Simulate stdinIsATTY")

	return executeTemplateCmd
}

func (c *Config) runExecuteTemplateCmd(cmd *cobra.Command, args []string) error {
	options := []chezmoi.SourceStateOption{
		chezmoi.WithTemplateDataOnly(true),
	}
	if c.executeTemplate.init {
		options = append(options, chezmoi.WithReadTemplateData(false))
	}
	sourceState, err := c.newSourceState(cmd.Context(), options...)
	if err != nil {
		return err
	}

	promptBool := make(map[string]bool)
	for key, valueStr := range c.executeTemplate.promptBool {
		value, err := parseBool(valueStr)
		if err != nil {
			return err
		}
		promptBool[key] = value
	}
	if c.executeTemplate.init {
		initTemplateFuncs := map[string]interface{}{
			"promptBool": func(prompt string, args ...bool) bool {
				switch len(args) {
				case 0:
					return promptBool[prompt]
				case 1:
					if value, ok := promptBool[prompt]; ok {
						return value
					}
					return args[0]
				default:
					err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
					raiseTemplateError(err)
					return false
				}
			},
			"promptInt": func(prompt string, args ...int) int {
				switch len(args) {
				case 0:
					return c.executeTemplate.promptInt[prompt]
				case 1:
					if value, ok := c.executeTemplate.promptInt[prompt]; ok {
						return value
					}
					return args[0]
				default:
					err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
					raiseTemplateError(err)
					return 0
				}
			},
			"promptString": func(prompt string, args ...string) string {
				switch len(args) {
				case 0:
					if value, ok := c.executeTemplate.promptString[prompt]; ok {
						return value
					}
					return prompt
				case 1:
					if value, ok := c.executeTemplate.promptString[prompt]; ok {
						return value
					}
					return args[0]
				default:
					err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
					raiseTemplateError(err)
					return ""
				}
			},
			"stdinIsATTY": func() bool {
				return c.executeTemplate.stdinIsATTY
			},
			"writeToStdout": c.writeToStdout,
		}
		for _, releaseTag := range build.Default.ReleaseTags {
			if releaseTag == "go1.17" {
				initTemplateFuncs["exit"] = c.exitInitTemplateFunc
				break
			}
		}
		chezmoi.RecursiveMerge(c.templateFuncs, initTemplateFuncs)
	}

	if len(args) == 0 {
		data, err := io.ReadAll(c.stdin)
		if err != nil {
			return err
		}
		output, err := sourceState.ExecuteTemplateData("stdin", data)
		if err != nil {
			return err
		}
		return c.writeOutput(output)
	}

	output := strings.Builder{}
	for i, arg := range args {
		result, err := sourceState.ExecuteTemplateData("arg"+strconv.Itoa(i+1), []byte(arg))
		if err != nil {
			return err
		}
		if _, err := output.Write(result); err != nil {
			return err
		}
	}
	return c.writeOutputString(output.String())
}
