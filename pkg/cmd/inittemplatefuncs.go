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
	switch len(args) {
	case 0:
		value, err := parseBool(c.promptStringInitTemplateFunc(prompt))
		if err != nil {
			panic(err)
		}
		return value
	case 1:
		prompt += " (default " + strconv.FormatBool(args[0]) + ")"
		valueStr := c.promptStringInitTemplateFunc(prompt)
		if valueStr == "" {
			return args[0]
		}
		value, err := parseBool(valueStr)
		if err != nil {
			panic(err)
		}
		return value
	default:
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}
}

func (c *Config) promptIntInitTemplateFunc(prompt string, args ...int64) int64 {
	switch len(args) {
	case 0:
		value, err := strconv.ParseInt(c.promptStringInitTemplateFunc(prompt), 10, 64)
		if err != nil {
			panic(err)
		}
		return value
	case 1:
		promptStr := prompt + " (default " + strconv.FormatInt(args[0], 10) + ")"
		valueStr := c.promptStringInitTemplateFunc(promptStr)
		if valueStr == "" {
			return args[0]
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			panic(err)
		}
		return value
	default:
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
		panic(err)
	}
}

func (c *Config) promptStringInitTemplateFunc(prompt string, args ...string) string {
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
		err := fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
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
