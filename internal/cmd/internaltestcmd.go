package cmd

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
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
		Use:   "prompt-bool",
		Args:  cobra.MinimumNArgs(1),
		Short: "Run promptBool",
		RunE:  c.runInternalTestPromptBoolCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	internalTestCmd.AddCommand(internalTestPromptBoolCmd)

	internalTestPromptChoiceCmd := &cobra.Command{
		Use:   "prompt-choice",
		Args:  cobra.MinimumNArgs(2),
		Short: "Run promptChoice",
		RunE:  c.runInternalTestPromptChoiceCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	internalTestCmd.AddCommand(internalTestPromptChoiceCmd)
	internalTestPromptIntCmd := &cobra.Command{
		Use:   "prompt-int",
		Args:  cobra.MinimumNArgs(1),
		Short: "Run promptInt",
		RunE:  c.runInternalTestPromptIntCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	internalTestCmd.AddCommand(internalTestPromptIntCmd)

	internalTestPromptMultichoiceCmd := &cobra.Command{
		Use:   "prompt-multichoice",
		Args:  cobra.MinimumNArgs(2),
		Short: "Run promptMultichoice",
		RunE:  c.runInternalTestPromptMultichoiceCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	internalTestCmd.AddCommand(internalTestPromptMultichoiceCmd)

	internalTestPromptStringCmd := &cobra.Command{
		Use:   "prompt-string",
		Args:  cobra.MinimumNArgs(1),
		Short: "Run promptString",
		RunE:  c.runInternalTestPromptStringCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
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
	}
	internalTestCmd.AddCommand(internalTestReadPasswordCmd)

	return internalTestCmd
}

func (c *Config) runInternalTestPromptBoolCmd(cmd *cobra.Command, args []string) error {
	boolArgs := make([]bool, 0, len(args)-1)
	for _, arg := range args[1:] {
		boolArg, err := chezmoi.ParseBool(arg)
		if err != nil {
			return err
		}
		boolArgs = append(boolArgs, boolArg)
	}
	value, err := c.promptBool(args[0], boolArgs...)
	if err != nil {
		return err
	}
	return c.writeOutputString(strconv.FormatBool(value) + "\n")
}

func (c *Config) runInternalTestPromptChoiceCmd(cmd *cobra.Command, args []string) error {
	value, err := c.promptChoice(args[0], strings.Split(args[1], ","), args[2:]...)
	if err != nil {
		return err
	}
	return c.writeOutputString(value + "\n")
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
		if err := c.writeOutputString(entry + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) runInternalTestPromptIntCmd(cmd *cobra.Command, args []string) error {
	int64Args := make([]int64, 0, len(args)-1)
	for _, arg := range args[1:] {
		int64Arg, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return err
		}
		int64Args = append(int64Args, int64Arg)
	}
	value, err := c.promptInt(args[0], int64Args...)
	if err != nil {
		return err
	}
	return c.writeOutputString(strconv.FormatInt(value, 10) + "\n")
}

func (c *Config) runInternalTestPromptStringCmd(cmd *cobra.Command, args []string) error {
	value, err := c.promptString(args[0], args[1:]...)
	if err != nil {
		return err
	}
	return c.writeOutputString(value + "\n")
}

func (c *Config) runInternalTestReadPasswordCmd(cmd *cobra.Command, args []string) error {
	password, err := c.readPassword("Password? ", "password")
	if err != nil {
		return err
	}
	return c.writeOutputString(password + "\n")
}
