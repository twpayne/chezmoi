package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func (c *Config) promptBool(field string, args ...bool) bool {
	switch len(args) {
	case 0:
		value, err := parseBool(c.promptString(field))
		if err != nil {
			returnTemplateError(err)
			return false
		}
		return value
	case 1:
		promptStr := field + " (default " + strconv.FormatBool(args[0]) + ")"
		valueStr := c.promptString(promptStr)
		if valueStr == "" {
			return args[0]
		}
		value, err := parseBool(valueStr)
		if err != nil {
			returnTemplateError(err)
			return false
		}
		return value
	default:
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		returnTemplateError(err)
		return false
	}
}

func (c *Config) promptInt(field string, args ...int64) int64 {
	switch len(args) {
	case 0:
		value, err := strconv.ParseInt(c.promptString(field), 10, 64)
		if err != nil {
			returnTemplateError(err)
			return 0
		}
		return value
	case 1:
		promptStr := field + " (default " + strconv.FormatInt(args[0], 10) + ")"
		valueStr := c.promptString(promptStr)
		if valueStr == "" {
			return args[0]
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			returnTemplateError(err)
			return 0
		}
		return value
	default:
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		returnTemplateError(err)
		return 0
	}
}

func (c *Config) promptString(prompt string, args ...string) string {
	switch len(args) {
	case 0:
		value, err := c.readLine(prompt + "? ")
		if err != nil {
			returnTemplateError(err)
			return ""
		}
		return strings.TrimSpace(value)
	case 1:
		defaultStr := strings.TrimSpace(args[0])
		promptStr := prompt + " (default " + strconv.Quote(defaultStr) + ")? "
		switch value, err := c.readLine(promptStr); {
		case err != nil:
			returnTemplateError(err)
			return ""
		case value == "":
			return defaultStr
		default:
			return strings.TrimSpace(value)
		}
	default:
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		returnTemplateError(err)
		return ""
	}
}

func (c *Config) stdinIsATTY() bool {
	file, ok := c.stdin.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func (c *Config) writeToStdout(args ...string) string {
	for _, arg := range args {
		if _, err := c.stdout.Write([]byte(arg)); err != nil {
			returnTemplateError(err)
		}
	}
	return ""
}
