package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type interactiveTemplateFuncsConfig struct {
	forcePromptOnce   bool
	promptBool        map[string]string
	promptChoice      map[string]string
	promptDefaults    bool
	promptInt         map[string]int
	promptMultichoice map[string]string
	promptString      map[string]string
}

func (c *Config) addInteractiveTemplateFuncFlags(flags *pflag.FlagSet) {
	flags.BoolVar(
		&c.interactiveTemplateFuncs.forcePromptOnce,
		"prompt",
		c.interactiveTemplateFuncs.forcePromptOnce,
		"Force prompt*Once template functions to prompt",
	)
	flags.BoolVar(
		&c.interactiveTemplateFuncs.promptDefaults,
		"promptDefaults",
		c.interactiveTemplateFuncs.promptDefaults,
		"Make prompt functions return default values",
	)
	flags.StringToStringVar(
		&c.interactiveTemplateFuncs.promptBool,
		"promptBool",
		c.interactiveTemplateFuncs.promptBool,
		"Populate promptBool",
	)
	flags.StringToStringVar(
		&c.interactiveTemplateFuncs.promptChoice,
		"promptChoice",
		c.interactiveTemplateFuncs.promptChoice,
		"Populate promptChoice",
	)
	flags.StringToStringVar(
		&c.interactiveTemplateFuncs.promptMultichoice,
		"promptMultichoice",
		c.interactiveTemplateFuncs.promptMultichoice,
		"Populate promptMultichoice",
	)
	flags.StringToIntVar(
		&c.interactiveTemplateFuncs.promptInt,
		"promptInt",
		c.interactiveTemplateFuncs.promptInt,
		"Populate promptInt",
	)
	flags.StringToStringVar(
		&c.interactiveTemplateFuncs.promptString,
		"promptString",
		c.interactiveTemplateFuncs.promptString,
		"Populate promptString",
	)
}

func (c *Config) promptBoolInteractiveTemplateFunc(prompt string, args ...bool) bool {
	if len(args) > 1 {
		panic(fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1))
	}

	if valueStr, ok := c.interactiveTemplateFuncs.promptBool[prompt]; ok {
		return mustValue(chezmoi.ParseBool(valueStr))
	}

	return mustValue(c.promptBool(prompt, args...))
}

func (c *Config) promptBoolOnceInteractiveTemplateFunc(m map[string]any, path any, prompt string, args ...bool) bool {
	if len(args) > 1 {
		panic(fmt.Errorf("want 3 or 4 arguments, got %d", len(args)+2))
	}

	nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
	if !c.interactiveTemplateFuncs.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			switch value := value.(type) {
			case bool:
				return value
			case string:
				if boolValue, err := chezmoi.ParseBool(value); err == nil {
					return boolValue
				}
			}
		}
	}

	return c.promptBoolInteractiveTemplateFunc(prompt, args...)
}

func (c *Config) promptChoiceInteractiveTemplateFunc(prompt string, choices any, args ...string) string {
	if len(args) > 1 {
		panic(fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2))
	}

	if valueStr, ok := c.interactiveTemplateFuncs.promptChoice[prompt]; ok {
		return valueStr
	}

	choiceStrs := mustValue(anyToStringSlice(choices))
	return mustValue(c.promptChoice(prompt, choiceStrs, args...))
}

func (c *Config) promptChoiceOnceInteractiveTemplateFunc(
	m map[string]any,
	path any,
	prompt string,
	choices any,
	args ...string,
) string {
	if len(args) > 1 {
		panic(fmt.Errorf("want 4 or 5 arguments, got %d", len(args)+4))
	}

	nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
	if !c.interactiveTemplateFuncs.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			if valueStr, ok := value.(string); ok {
				return valueStr
			}
		}
	}

	return c.promptChoiceInteractiveTemplateFunc(prompt, choices, args...)
}

func (c *Config) promptMultichoiceInteractiveTemplateFunc(prompt string, choices any, args ...any) []string {
	if len(args) > 1 {
		panic(fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2))
	}

	if valueStr, ok := c.interactiveTemplateFuncs.promptMultichoice[prompt]; ok {
		return strings.Split(valueStr, "/")
	}

	choiceStrs := mustValue(anyToStringSlice(choices))

	var defaultValue *[]string

	if len(args) == 1 && args[0] != nil {
		values := mustValue(anyToStringSlice(args[0]))
		defaultValue = &values
	}

	return mustValue(c.promptMultichoice(prompt, choiceStrs, defaultValue))
}

func (c *Config) promptMultichoiceOnceInteractiveTemplateFunc(
	m map[string]any,
	path any,
	prompt string,
	choices any,
	args ...any,
) []string {
	if len(args) > 1 {
		panic(fmt.Errorf("want 4 or 5 arguments, got %d", len(args)+4))
	}

	nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
	if !c.interactiveTemplateFuncs.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			return mustValue(anyToStringSlice(value))
		}
	}

	return c.promptMultichoiceInteractiveTemplateFunc(prompt, choices, args...)
}

func (c *Config) promptIntInteractiveTemplateFunc(prompt string, args ...int64) int64 {
	if len(args) > 1 {
		panic(fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1))
	}

	if value, ok := c.interactiveTemplateFuncs.promptInt[prompt]; ok {
		return int64(value)
	}

	return mustValue(c.promptInt(prompt, args...))
}

func (c *Config) promptIntOnceInteractiveTemplateFunc(m map[string]any, path any, prompt string, args ...int64) int64 {
	if len(args) > 1 {
		panic(fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2))
	}

	nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
	if !c.interactiveTemplateFuncs.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			if intValue, ok := value.(int64); ok {
				return intValue
			}
		}
	}

	return c.promptIntInteractiveTemplateFunc(prompt, args...)
}

func (c *Config) promptStringInteractiveTemplateFunc(prompt string, args ...string) string {
	if len(args) > 1 {
		panic(fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1))
	}

	if value, ok := c.interactiveTemplateFuncs.promptString[prompt]; ok {
		return value
	}

	return mustValue(c.promptString(prompt, args...))
}

func (c *Config) promptStringOnceInteractiveTemplateFunc(m map[string]any, path any, prompt string, args ...string) string {
	if len(args) > 1 {
		panic(fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2))
	}

	nestedMap, lastKey := mustValues(nestedMapAtPath(m, path))
	if !c.interactiveTemplateFuncs.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			if stringValue, ok := value.(string); ok {
				return stringValue
			}
		}
	}

	return c.promptStringInteractiveTemplateFunc(prompt, args...)
}
