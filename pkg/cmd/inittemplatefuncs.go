package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
		value, err := parseBool(valueStr)
		if err != nil {
			panic(err)
		}
		return value
	}

	switch len(args) {
	case 0:
		value, err := parseBool(c.promptString(prompt))
		if err != nil {
			panic(err)
		}
		return value
	case 1:
		prompt += " (default " + strconv.FormatBool(args[0]) + ")"
		valueStr := c.promptString(prompt)
		if valueStr == "" {
			return args[0]
		}
		value, err := parseBool(valueStr)
		if err != nil {
			panic(err)
		}
		return value
	default:
		panic("unreachable")
	}
}

func (c *Config) promptBoolOnceInitTemplateFunc(m map[string]any, key, prompt string, args ...bool) bool {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	if !c.init.forcePromptOnce {
		if value, ok := m[key]; ok {
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

	switch len(args) {
	case 0:
		value, err := strconv.ParseInt(c.promptString(prompt), 10, 64)
		if err != nil {
			panic(err)
		}
		return value
	case 1:
		promptStr := prompt + " (default " + strconv.FormatInt(args[0], 10) + ")"
		valueStr := c.promptString(promptStr)
		if valueStr == "" {
			return args[0]
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			panic(err)
		}
		return value
	default:
		panic("unreachable")
	}
}

func (c *Config) promptIntOnceInitTemplateFunc(m map[string]any, key, prompt string, args ...int64) int64 {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	if !c.init.forcePromptOnce {
		if value, ok := m[key]; ok {
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

	return c.promptString(prompt, args...)
}

func (c *Config) promptStringOnceInitTemplateFunc(m map[string]any, key, prompt string, args ...string) string {
	if len(args) > 1 {
		err := fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
		panic(err)
	}

	if !c.init.forcePromptOnce {
		if value, ok := m[key]; ok {
			if stringValue, ok := value.(string); ok {
				return stringValue
			}
		}
	}

	return c.promptStringInitTemplateFunc(prompt, args...)
}

func (c *Config) promptString(prompt string, args ...string) string {
	switch len(args) {
	case 0:
		value, err := c.readLine(prompt + "? ")
		if err != nil {
			panic(err)
		}
		return strings.TrimSpace(value)
	case 1:
		defaultStr := strings.TrimSpace(args[0])
		promptStr := prompt + " (default " + strconv.Quote(defaultStr) + ")? "
		switch value, err := c.readLine(promptStr); {
		case err != nil:
			panic(err)
		case value == "":
			return defaultStr
		default:
			return strings.TrimSpace(value)
		}
	default:
		err := fmt.Errorf("want 0 or 1 arguments, got %d", len(args))
		panic(err)
	}
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
