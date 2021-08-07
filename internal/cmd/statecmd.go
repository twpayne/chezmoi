package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type stateCmdConfig struct {
	data   stateDataCmdConfig
	delete stateDeleteCmdConfig
	dump   stateDumpCmdConfig
	get    stateGetCmdConfig
	set    stateSetCmdConfig
}

type stateDataCmdConfig struct {
	format writeDataFormat
}

type stateDeleteCmdConfig struct {
	bucket string
	key    string
}

type stateDumpCmdConfig struct {
	format writeDataFormat
}

type stateGetCmdConfig struct {
	bucket string
	key    string
}

type stateSetCmdConfig struct {
	bucket string
	key    string
	value  string
}

func (c *Config) newStateCmd() *cobra.Command {
	stateCmd := &cobra.Command{
		Use:     "state",
		Short:   "Manipulate the persistent state",
		Long:    mustLongHelp("state"),
		Example: example("state"),
	}

	stateDataCmd := &cobra.Command{
		Use:   "data",
		Short: "Print the raw data in the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateDataCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadOnly,
		},
	}
	stateDataPersistentFlags := stateDataCmd.PersistentFlags()
	stateDataPersistentFlags.VarP(&c.state.data.format, "format", "f", "format")
	stateCmd.AddCommand(stateDataCmd)

	stateDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a value from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateDeleteCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadWrite,
		},
	}
	stateDeletePersistentFlags := stateDeleteCmd.PersistentFlags()
	stateDeletePersistentFlags.StringVar(&c.state.delete.bucket, "bucket", c.state.delete.bucket, "bucket")
	stateDeletePersistentFlags.StringVar(&c.state.delete.key, "key", c.state.delete.key, "key")
	stateCmd.AddCommand(stateDeleteCmd)

	stateDumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Generate a dump of the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateDumpCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadOnly,
		},
	}
	stateDumpPersistentFlags := stateDumpCmd.PersistentFlags()
	stateDumpPersistentFlags.VarP(&c.state.dump.format, "format", "f", "format")
	stateCmd.AddCommand(stateDumpCmd)

	stateGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a value from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateGetCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadOnly,
		},
	}
	stateGetPersistentFlags := stateGetCmd.PersistentFlags()
	stateGetPersistentFlags.StringVar(&c.state.get.bucket, "bucket", c.state.get.bucket, "bucket")
	stateGetPersistentFlags.StringVar(&c.state.get.key, "key", c.state.get.key, "key")
	stateCmd.AddCommand(stateGetCmd)

	stateResetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateResetCmd,
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
		},
	}
	stateCmd.AddCommand(stateResetCmd)

	stateSetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set a value from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateSetCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadWrite,
		},
	}
	stateSetPersistentFlags := stateSetCmd.PersistentFlags()
	stateSetPersistentFlags.StringVar(&c.state.set.bucket, "bucket", c.state.set.bucket, "bucket")
	stateSetPersistentFlags.StringVar(&c.state.set.key, "key", c.state.set.key, "key")
	stateSetPersistentFlags.StringVar(&c.state.set.value, "value", c.state.set.value, "value")
	stateCmd.AddCommand(stateSetCmd)

	return stateCmd
}

func (c *Config) runStateDataCmd(cmd *cobra.Command, args []string) error {
	data, err := c.persistentState.Data()
	if err != nil {
		return err
	}
	return c.marshal(c.state.data.format, data)
}

func (c *Config) runStateDeleteCmd(cmd *cobra.Command, args []string) error {
	return c.persistentState.Delete([]byte(c.state.delete.bucket), []byte(c.state.delete.key))
}

func (c *Config) runStateDumpCmd(cmd *cobra.Command, args []string) error {
	data, err := chezmoi.PersistentStateData(c.persistentState)
	if err != nil {
		return err
	}
	return c.marshal(c.state.dump.format, data)
}

func (c *Config) runStateGetCmd(cmd *cobra.Command, args []string) error {
	value, err := c.persistentState.Get([]byte(c.state.get.bucket), []byte(c.state.get.key))
	if err != nil {
		return err
	}
	return c.writeOutput(value)
}

func (c *Config) runStateResetCmd(cmd *cobra.Command, args []string) error {
	persistentStateFileAbsPath, err := c.persistentStateFile()
	if err != nil {
		return err
	}
	switch _, err := c.destSystem.Stat(persistentStateFileAbsPath); {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	}
	if !c.force {
		switch choice, err := c.promptChoice(fmt.Sprintf("Remove %s", persistentStateFileAbsPath), []string{"yes", "no"}); {
		case err != nil:
			return err
		case choice == "yes":
		case choice == "no":
			return nil
		}
	}
	return c.destSystem.RemoveAll(persistentStateFileAbsPath)
}

func (c *Config) runStateSetCmd(cmd *cobra.Command, args []string) error {
	return c.persistentState.Set([]byte(c.state.set.bucket), []byte(c.state.set.key), []byte(c.state.set.value))
}
