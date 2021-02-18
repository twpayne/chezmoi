package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type stateCmdConfig struct {
	dump stateDumpCmdConfig
}

type stateDumpCmdConfig struct {
	format string
}

func (c *Config) newStateCmd() *cobra.Command {
	stateCmd := &cobra.Command{
		Use:     "state",
		Short:   "Manipulate the persistent state",
		Long:    mustLongHelp("state"),
		Example: example("state"),
	}

	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Generate a dump of the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateDumpCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadOnly,
		},
	}

	persistentFlags := dumpCmd.PersistentFlags()
	persistentFlags.StringVarP(&c.state.dump.format, "format", "f", c.state.dump.format, "format (json or yaml)")

	stateCmd.AddCommand(dumpCmd)

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateResetCmd,
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
		},
	}
	stateCmd.AddCommand(resetCmd)

	return stateCmd
}

func (c *Config) runStateDumpCmd(cmd *cobra.Command, args []string) error {
	data, err := chezmoi.PersistentStateData(c.persistentState)
	if err != nil {
		return err
	}
	return c.marshal(c.state.dump.format, data)
}

func (c *Config) runStateResetCmd(cmd *cobra.Command, args []string) error {
	path := c.persistentStateFile()
	_, err := c.destSystem.Stat(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if !c.force {
		switch choice, err := c.promptChoice(fmt.Sprintf("Remove %s", path), []string{"yes", "no"}); {
		case err != nil:
			return err
		case choice == "yes":
		case choice == "no":
			return nil
		}
	}
	return c.destSystem.RemoveAll(path)
}
