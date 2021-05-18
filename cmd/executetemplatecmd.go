package cmd

import (
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
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
		RunE:    c.makeRunEWithSourceState(c.runExecuteTemplateCmd),
	}

	flags := executeTemplateCmd.Flags()
	flags.BoolVarP(&c.executeTemplate.init, "init", "i", c.executeTemplate.init, "simulate chezmoi init")
	flags.StringToStringVar(&c.executeTemplate.promptBool, "promptBool", c.executeTemplate.promptBool, "simulate promptBool")
	flags.StringToIntVar(&c.executeTemplate.promptInt, "promptInt", c.executeTemplate.promptInt, "simulate promptInt")
	flags.StringToStringVarP(&c.executeTemplate.promptString, "promptString", "p", c.executeTemplate.promptString, "simulate promptString")
	flags.BoolVar(&c.executeTemplate.stdinIsATTY, "stdinisatty", c.executeTemplate.stdinIsATTY, "simulate stdinIsATTY")

	return executeTemplateCmd
}

func (c *Config) runExecuteTemplateCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	promptBool := make(map[string]bool)
	for key, valueStr := range c.executeTemplate.promptBool {
		value, err := parseBool(valueStr)
		if err != nil {
			return err
		}
		promptBool[key] = value
	}
	if c.executeTemplate.init {
		chezmoi.RecursiveMerge(c.templateFuncs, map[string]interface{}{
			"promptBool": func(prompt string) bool {
				return promptBool[prompt]
			},
			"promptInt": func(prompt string) int {
				return c.executeTemplate.promptInt[prompt]
			},
			"promptString": func(prompt string) string {
				if value, ok := c.executeTemplate.promptString[prompt]; ok {
					return value
				}
				return prompt
			},
			"stdinIsATTY": func() bool {
				return c.executeTemplate.stdinIsATTY
			},
			"writeToStdout": c.writeToStdout,
		})
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
