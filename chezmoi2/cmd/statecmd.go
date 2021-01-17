package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

func (c *Config) newStateCmd() *cobra.Command {
	stateCmd := &cobra.Command{
		Use:   "state",
		Short: "Manipulate the persistent state",
	}

	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Generate a dump of the persistent state",
		// Long: mustLongHelp("state", "dump"), // FIXME
		// Example: example("state", "dump"), // FIXME
		Args: cobra.NoArgs,
		RunE: c.runStateDataCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadOnly,
		},
	}
	stateCmd.AddCommand(dumpCmd)

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the persistent state",
		// Long: mustLongHelp("state", "reset"), // FIXME
		// Example: example("state", "reset"), // FIXME
		Args: cobra.NoArgs,
		RunE: c.runStateResetCmd,
	}
	stateCmd.AddCommand(resetCmd)

	return stateCmd
}

func (c *Config) runStateDataCmd(cmd *cobra.Command, args []string) error {
	data, err := chezmoi.PersistentStateData(c.persistentState)
	if err != nil {
		return err
	}
	return c.marshal(data)
}

func (c *Config) runStateResetCmd(cmd *cobra.Command, args []string) error {
	path := c.persistentStateFile()
	_, err := c.baseSystem.Stat(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if !c.force {
		choice, err := c.prompt(fmt.Sprintf("Remove %s", path), "yn")
		if err != nil {
			return err
		}
		if choice == 'n' {
			return nil
		}
	}
	return c.baseSystem.RemoveAll(path)
}
