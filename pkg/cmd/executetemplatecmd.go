package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type executeTemplateCmdConfig struct {
	init            bool
	promptBool      map[string]string
	promptInt       map[string]int
	promptString    map[string]string
	stdinIsATTY     bool
	templateOptions chezmoi.TemplateOptions
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
	flags.StringVar(&c.executeTemplate.templateOptions.LeftDelimiter, "left-delimiter", c.executeTemplate.templateOptions.LeftDelimiter, "Set left template delimiter")     //nolint:lll
	flags.StringVar(&c.executeTemplate.templateOptions.RightDelimiter, "right-delimiter", c.executeTemplate.templateOptions.RightDelimiter, "Set right template delimiter") //nolint:lll

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
		value, err := chezmoi.ParseBool(valueStr)
		if err != nil {
			return err
		}
		promptBool[key] = value
	}
	if c.executeTemplate.init {
		promptBoolInitTemplateFunc := func(prompt string, args ...bool) bool {
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
				panic(err)
			}
		}

		promptBoolOnceInitTemplateFunc := func(m map[string]any, key, field string, args ...bool) bool {
			if value, ok := m[key]; ok {
				if boolValue, ok := value.(bool); ok {
					return boolValue
				}
			}
			return promptBoolInitTemplateFunc(field, args...)
		}

		promptIntInitTemplateFunc := func(prompt string, args ...int64) int64 {
			switch len(args) {
			case 0:
				return int64(c.executeTemplate.promptInt[prompt])
			case 1:
				if value, ok := c.executeTemplate.promptInt[prompt]; ok {
					return int64(value)
				}
				return args[0]
			default:
				err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
				panic(err)
			}
		}

		promptIntOnceInitTemplateFunc := func(m map[string]any, key, field string, args ...int64) int64 {
			if value, ok := m[key]; ok {
				if intValue, ok := value.(int64); ok {
					return intValue
				}
			}
			return promptIntInitTemplateFunc(field, args...)
		}

		promptStringInitTemplateFunc := func(prompt string, args ...string) string {
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
				panic(err)
			}
		}

		promptStringOnceInitTemplateFunc := func(m map[string]any, key, field string, args ...string) string {
			if value, ok := m[key]; ok {
				if stringValue, ok := value.(string); ok {
					return stringValue
				}
			}
			return promptStringInitTemplateFunc(field, args...)
		}

		stdinIsATTYInitTemplateFunc := func() bool {
			return c.executeTemplate.stdinIsATTY
		}

		initTemplateFuncs := map[string]any{
			"exit":             c.exitInitTemplateFunc,
			"promptBool":       promptBoolInitTemplateFunc,
			"promptBoolOnce":   promptBoolOnceInitTemplateFunc,
			"promptInt":        promptIntInitTemplateFunc,
			"promptIntOnce":    promptIntOnceInitTemplateFunc,
			"promptString":     promptStringInitTemplateFunc,
			"promptStringOnce": promptStringOnceInitTemplateFunc,
			"stdinIsATTY":      stdinIsATTYInitTemplateFunc,
			"writeToStdout":    c.writeToStdout,
		}

		chezmoi.RecursiveMerge(c.templateFuncs, initTemplateFuncs)
	}

	if len(args) == 0 {
		data, err := io.ReadAll(c.stdin)
		if err != nil {
			return err
		}
		output, err := sourceState.ExecuteTemplateData(chezmoi.ExecuteTemplateDataOptions{
			Name:            "stdin",
			Data:            data,
			TemplateOptions: c.executeTemplate.templateOptions,
		})
		if err != nil {
			return err
		}
		return c.writeOutput(output)
	}

	output := strings.Builder{}
	for i, arg := range args {
		result, err := sourceState.ExecuteTemplateData(chezmoi.ExecuteTemplateDataOptions{
			Name:            "arg" + strconv.Itoa(i+1),
			Data:            []byte(arg),
			TemplateOptions: c.executeTemplate.templateOptions,
		})
		if err != nil {
			return err
		}
		if _, err := output.Write(result); err != nil {
			return err
		}
	}
	return c.writeOutputString(output.String())
}
