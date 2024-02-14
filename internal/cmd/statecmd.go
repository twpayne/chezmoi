package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type stateCmdConfig struct {
	delete       stateDeleteCmdConfig
	deleteBucket stateDeleteBucketCmdConfig
	get          stateGetCmdConfig
	getBucket    stateGetBucketCmdConfig
	set          stateSetCmdConfig
}

type stateDeleteCmdConfig struct {
	bucket string
	key    string
}

type stateDeleteBucketCmdConfig struct {
	bucket string
}

type stateGetCmdConfig struct {
	bucket string
	key    string
}

type stateGetBucketCmdConfig struct {
	bucket string
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
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}
	stateDataCmd.Flags().VarP(&c.Format, "format", "f", "Output format")
	if err := stateDataCmd.RegisterFlagCompletionFunc("format", writeDataFormatFlagCompletionFunc); err != nil {
		panic(err)
	}
	stateCmd.AddCommand(stateDataCmd)

	stateDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a value from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateDeleteCmd,
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	stateDeleteCmd.Flags().StringVar(&c.state.delete.bucket, "bucket", c.state.delete.bucket, "Bucket")
	stateDeleteCmd.Flags().StringVar(&c.state.delete.key, "key", c.state.delete.key, "Key")
	stateCmd.AddCommand(stateDeleteCmd)

	stateDeleteBucketCmd := &cobra.Command{
		Use:   "delete-bucket",
		Short: "Delete a bucket from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateDeleteBucketCmd,
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	stateDeleteBucketCmd.Flags().StringVar(&c.state.deleteBucket.bucket, "bucket", c.state.deleteBucket.bucket, "Bucket")
	stateCmd.AddCommand(stateDeleteBucketCmd)

	stateDumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Generate a dump of the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateDumpCmd,
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}
	stateDumpCmd.Flags().VarP(&c.Format, "format", "f", "Output format")
	if err := stateDumpCmd.RegisterFlagCompletionFunc("format", writeDataFormatFlagCompletionFunc); err != nil {
		panic(err)
	}
	stateCmd.AddCommand(stateDumpCmd)

	stateGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a value from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateGetCmd,
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}
	stateGetCmd.Flags().StringVar(&c.state.get.bucket, "bucket", c.state.get.bucket, "Bucket")
	stateGetCmd.Flags().StringVar(&c.state.get.key, "key", c.state.get.key, "Key")
	stateCmd.AddCommand(stateGetCmd)

	stateGetBucketCmd := &cobra.Command{
		Use:   "get-bucket",
		Short: "Get a bucket from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateGetBucketCmd,
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}
	stateGetBucketCmd.Flags().StringVar(&c.state.getBucket.bucket, "bucket", c.state.getBucket.bucket, "bucket")
	stateGetBucketCmd.Flags().VarP(&c.Format, "format", "f", "Output format")
	if err := stateGetBucketCmd.RegisterFlagCompletionFunc("format", writeDataFormatFlagCompletionFunc); err != nil {
		panic(err)
	}
	stateCmd.AddCommand(stateGetBucketCmd)

	stateResetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateResetCmd,
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
		),
	}
	stateCmd.AddCommand(stateResetCmd)

	stateSetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set a value from the persistent state",
		Args:  cobra.NoArgs,
		RunE:  c.runStateSetCmd,
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	stateSetCmd.Flags().StringVar(&c.state.set.bucket, "bucket", c.state.set.bucket, "Bucket")
	stateSetCmd.Flags().StringVar(&c.state.set.key, "key", c.state.set.key, "Key")
	stateSetCmd.Flags().StringVar(&c.state.set.value, "value", c.state.set.value, "Value")
	stateCmd.AddCommand(stateSetCmd)

	return stateCmd
}

func (c *Config) runStateDataCmd(cmd *cobra.Command, args []string) error {
	data, err := c.persistentState.Data()
	if err != nil {
		return err
	}
	return c.marshal(c.Format, data)
}

func (c *Config) runStateDeleteCmd(cmd *cobra.Command, args []string) error {
	return c.persistentState.Delete([]byte(c.state.delete.bucket), []byte(c.state.delete.key))
}

func (c *Config) runStateDeleteBucketCmd(cmd *cobra.Command, args []string) error {
	return c.persistentState.DeleteBucket([]byte(c.state.deleteBucket.bucket))
}

func (c *Config) runStateDumpCmd(cmd *cobra.Command, args []string) error {
	data, err := chezmoi.PersistentStateData(c.persistentState, map[string][]byte{
		"configState":              chezmoi.ConfigStateBucket,
		"entryState":               chezmoi.EntryStateBucket,
		"gitHubKeysState":          gitHubKeysStateBucket,
		"gitHubLatestReleaseState": gitHubLatestReleaseStateBucket,
		"gitHubReleasesState":      gitHubReleasesStateBucket,
		"gitHubTagsState":          gitHubTagsStateBucket,
		"gitRepoExternalState":     chezmoi.GitRepoExternalStateBucket,
		"scriptState":              chezmoi.ScriptStateBucket,
	})
	if err != nil {
		return err
	}
	return c.marshal(c.Format, data)
}

func (c *Config) runStateGetCmd(cmd *cobra.Command, args []string) error {
	value, err := c.persistentState.Get([]byte(c.state.get.bucket), []byte(c.state.get.key))
	if err != nil {
		return err
	}
	return c.writeOutput(value)
}

func (c *Config) runStateGetBucketCmd(cmd *cobra.Command, args []string) error {
	data, err := chezmoi.PersistentStateBucketData(c.persistentState, []byte(c.state.getBucket.bucket))
	if err != nil {
		return err
	}
	return c.marshal(c.Format, data)
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
		switch choice, err := c.promptChoice(fmt.Sprintf("Remove %s", persistentStateFileAbsPath), choicesYesNoQuit); {
		case err != nil:
			return err
		case choice == "yes":
		case choice == "no":
			fallthrough
		case choice == "quit":
			return nil
		}
	}
	return c.destSystem.RemoveAll(persistentStateFileAbsPath)
}

func (c *Config) runStateSetCmd(cmd *cobra.Command, args []string) error {
	return c.persistentState.Set([]byte(c.state.set.bucket), []byte(c.state.set.key), []byte(c.state.set.value))
}
