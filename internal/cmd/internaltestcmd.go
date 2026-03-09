package cmd

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

func (c *Config) newInternalTestCmd() *cobra.Command {
	internalTestCmd := &cobra.Command{
		Use:    "internal-test",
		Short:  "Expose functionality for testing",
		Hidden: true,
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}

	internalTestPromptBoolCmd := &cobra.Command{
		Use:   "prompt-bool prompt [default]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Run promptBool",
		RunE:  c.runInternalTestPromptBoolCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
		Example: "  chezmoi internal-test prompt-bool overwrite false",
	}
	internalTestCmd.AddCommand(internalTestPromptBoolCmd)

	internalTestPromptChoiceCmd := &cobra.Command{
		Use:   "prompt-choice prompt choices [default]",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Run promptChoice",
		RunE:  c.runInternalTestPromptChoiceCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
		Example: "  chezmoi internal-test prompt-choice color red,green,blue red",
	}
	internalTestCmd.AddCommand(internalTestPromptChoiceCmd)
	internalTestPromptIntCmd := &cobra.Command{
		Use:   "prompt-int prompt [default]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Run promptInt",
		RunE:  c.runInternalTestPromptIntCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
		Example: "  chezmoi internal-test prompt-int count 1",
	}
	internalTestCmd.AddCommand(internalTestPromptIntCmd)

	internalTestPromptMultichoiceCmd := &cobra.Command{
		Use:   "prompt-multichoice prompt choices [defaults]",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Run promptMultichoice",
		RunE:  c.runInternalTestPromptMultichoiceCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
		Example: "  chezmoi internal-test prompt-multichoice days mon,tue,wed,thu,fri,sat,sun sat,sun",
	}
	internalTestCmd.AddCommand(internalTestPromptMultichoiceCmd)

	internalTestPromptStringCmd := &cobra.Command{
		Use:   "prompt-string prompt [default]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Run promptString",
		RunE:  c.runInternalTestPromptStringCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
		Example: "  chezmoi internal-test prompt-string username root",
	}
	internalTestCmd.AddCommand(internalTestPromptStringCmd)

	internalTestReadPasswordCmd := &cobra.Command{
		Use:   "read-password",
		Args:  cobra.NoArgs,
		Short: "Read a password",
		RunE:  c.runInternalTestReadPasswordCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
		Example: "  chezmoi internal-test read-password",
	}
	internalTestCmd.AddCommand(internalTestReadPasswordCmd)

	return internalTestCmd
}

func (c *Config) runInternalTestPromptBoolCmd(cmd *cobra.Command, args []string) error {
	boolArgs := make([]bool, len(args)-1)
	for i, arg := range args[1:] {
		boolArg, err := chezmoi.ParseBool(arg)
		if err != nil {
			return err
		}
		boolArgs[i] = boolArg
	}
	value, err := c.promptBool(args[0], boolArgs...)
	if err != nil {
		return err
	}
	return c.writeOutputString(strconv.FormatBool(value)+"\n", 0o666)
}

func (c *Config) runInternalTestPromptChoiceCmd(cmd *cobra.Command, args []string) error {
	value, err := c.promptChoice(args[0], strings.Split(args[1], ","), args[2:]...)
	if err != nil {
		return err
	}
	return c.writeOutputString(value+"\n", 0o666)
}

func (c *Config) runInternalTestPromptMultichoiceCmd(cmd *cobra.Command, args []string) error {
	var defaults *[]string
	if len(args) > 2 {
		values := strings.Split(args[2], ",")
		defaults = &values
	}

	value, err := c.promptMultichoice(args[0], strings.Split(args[1], ","), defaults)
	if err != nil {
		return err
	}

	for _, entry := range value {
		if err := c.writeOutputString(entry+"\n", 0o666); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) runInternalTestPromptIntCmd(cmd *cobra.Command, args []string) error {
	int64Args := make([]int64, len(args)-1)
	for i, arg := range args[1:] {
		int64Arg, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return err
		}
		int64Args[i] = int64Arg
	}
	value, err := c.promptInt(args[0], int64Args...)
	if err != nil {
		return err
	}
	return c.writeOutputString(strconv.FormatInt(value, 10)+"\n", 0o666)
}

func (c *Config) runInternalTestPromptStringCmd(cmd *cobra.Command, args []string) error {
	value, err := c.promptString(args[0], args[1:]...)
	if err != nil {
		return err
	}
	return c.writeOutputString(value+"\n", 0o666)
}

func (c *Config) runInternalTestReadPasswordCmd(cmd *cobra.Command, args []string) error {
	password, err := c.readPassword("Password? ", "password")
	if err != nil {
		return err
	}
	return c.writeOutputString(password+"\n", 0o666)
}
