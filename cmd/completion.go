package cmd

import (
	"bytes"
	"errors"

	"github.com/spf13/cobra"
)

type completionCmdConfig struct {
	output string
}

var completionCmd = &cobra.Command{
	Use:       "completion shell",
	Args:      cobra.ExactArgs(1),
	Short:     "Generate shell completion code for the specified shell (bash, fish, or zsh)",
	Long:      mustGetLongHelp("completion"),
	Example:   getExample("completion"),
	ValidArgs: []string{"bash", "fish", "zsh"},
	RunE:      config.runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)

	persistentFlags := completionCmd.PersistentFlags()
	persistentFlags.StringVarP(&config.completion.output, "output", "o", "", "output filename")
	panicOnError(completionCmd.MarkPersistentFlagFilename("output"))
}

func (c *Config) runCompletion(cmd *cobra.Command, args []string) error {
	output := &bytes.Buffer{}
	switch args[0] {
	case "bash":
		if err := rootCmd.GenBashCompletion(output); err != nil {
			return err
		}
	case "zsh":
		if err := rootCmd.GenZshCompletion(output); err != nil {
			return err
		}
	case "fish":
		if err := rootCmd.GenFishCompletion(output, true); err != nil {
			return err
		}
	default:
		return errors.New("unsupported shell")
	}

	if c.completion.output == "" {
		_, err := c.Stdout.Write(output.Bytes())
		return err
	}
	return c.fs.WriteFile(c.completion.output, output.Bytes(), 0o666)
}
