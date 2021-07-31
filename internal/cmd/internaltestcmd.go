package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newInternalTestCmd() *cobra.Command {
	internalTestCmd := &cobra.Command{
		Use:    "internal-test",
		Short:  "Expose functionality for testing",
		Hidden: true,
	}

	internalTestReadPasswordCmd := &cobra.Command{
		Use:   "read-password",
		Short: "Read a password",
		RunE:  c.runInternalTestReadPasswordCmd,
	}
	internalTestCmd.AddCommand(internalTestReadPasswordCmd)

	return internalTestCmd
}

func (c *Config) runInternalTestReadPasswordCmd(cmd *cobra.Command, args []string) error {
	password, err := c.readPassword("Password? ")
	if err != nil {
		return err
	}
	return c.writeOutputString(password + "\n")
}
