package cmd

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type executeTemplateCmdConfig struct {
	init              bool
	promptBool        map[string]string
	promptChoice      map[string]string
	promptInt         map[string]int
	promptMultichoice map[string]string
	promptString      map[string]string
	stdinIsATTY       bool
	templateOptions   chezmoi.TemplateOptions
	withStdin         bool
}

func (c *Config) newExecuteTemplateCmd() *cobra.Command {
	executeTemplateCmd := &cobra.Command{
		Use:     "execute-template [template]...",
		Short:   "Execute the given template(s)",
		Long:    mustLongHelp("execute-template"),
		Example: example("execute-template"),
		RunE:    c.runExecuteTemplateCmd,
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}

	executeTemplateCmd.Flags().BoolVarP(&c.executeTemplate.init, "init", "i", c.executeTemplate.init, "Simulate chezmoi init")
	executeTemplateCmd.Flags().
		StringToStringVar(&c.executeTemplate.promptBool, "promptBool", c.executeTemplate.promptBool, "Simulate promptBool")
	executeTemplateCmd.Flags().
		StringToStringVar(&c.executeTemplate.promptChoice, "promptChoice", c.executeTemplate.promptChoice, "Simulate promptChoice")
	executeTemplateCmd.Flags().
		StringToIntVar(&c.executeTemplate.promptInt, "promptInt", c.executeTemplate.promptInt, "Simulate promptInt")
	executeTemplateCmd.Flags().
		StringToStringVar(&c.executeTemplate.promptMultichoice, "promptMultichoice", c.executeTemplate.promptMultichoice, "Simulate promptMultichoice")
	executeTemplateCmd.Flags().
		StringToStringVarP(&c.executeTemplate.promptString, "promptString", "p", c.executeTemplate.promptString, "Simulate promptString")
	executeTemplateCmd.Flags().
		BoolVar(&c.executeTemplate.stdinIsATTY, "stdinisatty", c.executeTemplate.stdinIsATTY, "Simulate stdinIsATTY")
	executeTemplateCmd.Flags().
		StringVar(&c.executeTemplate.templateOptions.LeftDelimiter, "left-delimiter", c.executeTemplate.templateOptions.LeftDelimiter, "Set left template delimiter")
	executeTemplateCmd.Flags().
		StringVar(&c.executeTemplate.templateOptions.RightDelimiter, "right-delimiter", c.executeTemplate.templateOptions.RightDelimiter, "Set right template delimiter")
	executeTemplateCmd.Flags().
		BoolVar(&c.executeTemplate.withStdin, "with-stdin", c.executeTemplate.withStdin, "Set .chezmoi.stdin to the contents of the standard input")

	return executeTemplateCmd
}

func (c *Config) runExecuteTemplateCmd(cmd *cobra.Command, args []string) error {
	options := []chezmoi.SourceStateOption{
		chezmoi.WithTemplateDataOnly(true),
		chezmoi.WithReadTemplates(!c.executeTemplate.init),
	}
	if c.executeTemplate.init {
		options = append(options, chezmoi.WithReadTemplateData(false))
	}
	if c.executeTemplate.withStdin && len(args) > 0 {
		stdin, err := io.ReadAll(c.stdin)
		if err != nil {
			return err
		}
		options = append(options, chezmoi.WithPriorityTemplateData(map[string]any{
			"chezmoi": map[string]any{
				"stdin": string(stdin),
			},
		}))
	}
	sourceState, err := c.newSourceState(cmd.Context(), cmd, options...)
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
				panic(fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1))
			}
		}

		promptBoolOnceInitTemplateFunc := func(m map[string]any, path any, field string, args ...bool) bool {
			nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
			if value, ok := nestedMap[lastKey]; ok {
				if boolValue, ok := value.(bool); ok {
					return boolValue
				}
			}
			return promptBoolInitTemplateFunc(field, args...)
		}

		promptChoiceInitTemplateFunc := func(prompt string, choices any, args ...string) string {
			choiceStrs := mustValue(anyToStringSlice(choices))
			switch len(args) {
			case 0:
				if value, ok := c.executeTemplate.promptChoice[prompt]; ok {
					if !slices.Contains(choiceStrs, value) {
						panic(fmt.Errorf("%s: invalid choice", value))
					}
					return value
				}
				return prompt
			case 1:
				if value, ok := c.executeTemplate.promptChoice[prompt]; ok {
					if !slices.Contains(choiceStrs, value) {
						panic(fmt.Errorf("%s: invalid choice", value))
					}
					return value
				}
				return args[0]
			default:
				panic(fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+1))
			}
		}

		promptChoiceOnceInitTemplateFunc := func(m map[string]any, path any, prompt string, choices []any, args ...string) string {
			nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
			if value, ok := nestedMap[lastKey]; ok {
				if stringValue, ok := value.(string); ok {
					return stringValue
				}
			}
			return promptChoiceInitTemplateFunc(prompt, choices, args...)
		}

		promptMultichoiceInitTemplateFunc := func(prompt string, choices any, args ...any) []string {
			choiceStrs := mustValue(anyToStringSlice(choices))
			if value, ok := c.executeTemplate.promptMultichoice[prompt]; ok {
				values := strings.Split(value, "/")

				for _, v := range values {
					if !slices.Contains(choiceStrs, v) {
						panic(fmt.Errorf("%s: invalid choice", value))
					}
				}

				return values
			}

			if len(args) == 0 {
				return []string{prompt}
			}

			return mustValue(anyToStringSlice(args[0]))
		}

		promptMultichoiceOnceInitTemplateFunc := func(m map[string]any, path any, prompt string, choices []any, args ...any) []string {
			nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
			if value, ok := nestedMap[lastKey]; ok {
				if stringValue, ok := value.(string); ok {
					return strings.Split(stringValue, ",")
				}
			}
			return promptMultichoiceInitTemplateFunc(prompt, choices, args...)
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
				panic(fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1))
			}
		}

		promptIntOnceInitTemplateFunc := func(m map[string]any, path any, prompt string, args ...int64) int64 {
			nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
			if value, ok := nestedMap[lastKey]; ok {
				if intValue, ok := value.(int64); ok {
					return intValue
				}
			}
			return promptIntInitTemplateFunc(prompt, args...)
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
				panic(fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1))
			}
		}

		promptStringOnceInitTemplateFunc := func(m map[string]any, path any, prompt string, args ...string) string {
			nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
			if value, ok := nestedMap[lastKey]; ok {
				if stringValue, ok := value.(string); ok {
					return stringValue
				}
			}
			return promptStringInitTemplateFunc(prompt, args...)
		}

		stdinIsATTYInitTemplateFunc := func() bool {
			return c.executeTemplate.stdinIsATTY
		}

		initTemplateFuncs := map[string]any{
			"exit":                  c.exitInitTemplateFunc,
			"promptBool":            promptBoolInitTemplateFunc,
			"promptBoolOnce":        promptBoolOnceInitTemplateFunc,
			"promptChoice":          promptChoiceInitTemplateFunc,
			"promptChoiceOnce":      promptChoiceOnceInitTemplateFunc,
			"promptInt":             promptIntInitTemplateFunc,
			"promptIntOnce":         promptIntOnceInitTemplateFunc,
			"promptMultichoice":     promptMultichoiceInitTemplateFunc,
			"promptMultichoiceOnce": promptMultichoiceOnceInitTemplateFunc,
			"promptString":          promptStringInitTemplateFunc,
			"promptStringOnce":      promptStringOnceInitTemplateFunc,
			"stdinIsATTY":           stdinIsATTYInitTemplateFunc,
			"writeToStdout":         c.writeToStdout,
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
