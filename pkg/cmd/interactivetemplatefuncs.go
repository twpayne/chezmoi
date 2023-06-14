package cmd

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type interactiveTemplateFuncsConfig struct {
	forcePromptOnce bool
	promptBool      map[string]string
	promptInt       map[string]int
	promptString    map[string]string
}

func (c *Config) addInteractiveTemplateFuncFlags(flags *pflag.FlagSet) {
	flags.BoolVar(
		&c.interactiveTemplateFuncs.forcePromptOnce,
		"prompt",
		c.interactiveTemplateFuncs.forcePromptOnce,
		"Force prompt*Once template functions to prompt",
	)
	flags.StringToStringVar(
		&c.interactiveTemplateFuncs.promptBool,
		"promptBool",
		c.interactiveTemplateFuncs.promptBool,
		"Populate promptBool",
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
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}

	if valueStr, ok := c.interactiveTemplateFuncs.promptBool[prompt]; ok {
		value, err := chezmoi.ParseBool(valueStr)
		if err != nil {
			panic(err)
		}
		return value
	}

	value, err := c.promptBool(prompt, args...)
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptBoolOnceInteractiveTemplateFunc(
	m map[string]any,
	path any,
	prompt string,
	args ...bool,
) bool {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	nestedMap, lastKey, err := nestedMapAtPath(m, path)
	if err != nil {
		panic(err)
	}
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

func (c *Config) promptIntInteractiveTemplateFunc(prompt string, args ...int64) int64 {
	if len(args) > 1 {
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}

	if value, ok := c.interactiveTemplateFuncs.promptInt[prompt]; ok {
		return int64(value)
	}

	value, err := c.promptInt(prompt, args...)
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptIntOnceInteractiveTemplateFunc(
	m map[string]any,
	path any,
	prompt string,
	args ...int64,
) int64 {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	nestedMap, lastKey, err := nestedMapAtPath(m, path)
	if err != nil {
		panic(err)
	}
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
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}

	if value, ok := c.interactiveTemplateFuncs.promptString[prompt]; ok {
		return value
	}

	value, err := c.promptString(prompt, args...)
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptStringOnceInteractiveTemplateFunc(
	m map[string]any,
	path any,
	prompt string,
	args ...string,
) string {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	nestedMap, lastKey, err := nestedMapAtPath(m, path)
	if err != nil {
		panic(err)
	}
	if !c.interactiveTemplateFuncs.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			if stringValue, ok := value.(string); ok {
				return stringValue
			}
		}
	}

	return c.promptStringInteractiveTemplateFunc(prompt, args...)
}
