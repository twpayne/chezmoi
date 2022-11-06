package cmd

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (c *Config) exitInitTemplateFunc(code int) string {
	panic(chezmoi.ExitCodeError(code))
}

func (c *Config) promptBoolInitTemplateFunc(prompt string, args ...bool) bool {
	if len(args) > 1 {
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}

	if valueStr, ok := c.init.promptBool[prompt]; ok {
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

func (c *Config) promptBoolOnceInitTemplateFunc(m map[string]any, path any, prompt string, args ...bool) bool {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	nestedMap, lastKey, err := nestedMapAtPath(m, path)
	if err != nil {
		panic(err)
	}
	if !c.init.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			if boolValue, ok := value.(bool); ok {
				return boolValue
			}
		}
	}

	return c.promptBoolInitTemplateFunc(prompt, args...)
}

func (c *Config) promptIntInitTemplateFunc(prompt string, args ...int64) int64 {
	if len(args) > 1 {
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}

	if value, ok := c.init.promptInt[prompt]; ok {
		return int64(value)
	}

	value, err := c.promptInt(prompt, args...)
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptIntOnceInitTemplateFunc(m map[string]any, path any, prompt string, args ...int64) int64 {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	nestedMap, lastKey, err := nestedMapAtPath(m, path)
	if err != nil {
		panic(err)
	}
	if !c.init.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			if intValue, ok := value.(int64); ok {
				return intValue
			}
		}
	}

	return c.promptIntInitTemplateFunc(prompt, args...)
}

func (c *Config) promptStringInitTemplateFunc(prompt string, args ...string) string {
	if len(args) > 1 {
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}

	if value, ok := c.init.promptString[prompt]; ok {
		return value
	}

	value, err := c.promptString(prompt, args...)
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptStringOnceInitTemplateFunc(m map[string]any, path any, prompt string, args ...string) string {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	nestedMap, lastKey, err := nestedMapAtPath(m, path)
	if err != nil {
		panic(err)
	}
	if !c.init.forcePromptOnce {
		if value, ok := nestedMap[lastKey]; ok {
			if stringValue, ok := value.(string); ok {
				return stringValue
			}
		}
	}

	return c.promptStringInitTemplateFunc(prompt, args...)
}

func (c *Config) stdinIsATTYInitTemplateFunc() bool {
	file, ok := c.stdin.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func (c *Config) writeToStdout(args ...string) string {
	for _, arg := range args {
		if _, err := c.stdout.Write([]byte(arg)); err != nil {
			panic(err)
		}
	}
	return ""
}
