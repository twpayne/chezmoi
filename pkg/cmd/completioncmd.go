package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type completionCmdConfig struct {
	Custom bool `mapstructure:"custom"`
}

func (c *Config) newCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:       "completion shell",
		Short:     "Generate shell completion code",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "fish", "powershell", "zsh"},
		Long:      mustLongHelp("completion"),
		Example:   example("completion"),
		RunE:      c.runCompletionCmd,
		Annotations: map[string]string{
			doesNotRequireValidConfig: "true",
		},
	}

	return completionCmd
}

func (c *Config) runCompletionCmd(cmd *cobra.Command, args []string) error {
	builder := strings.Builder{}
	builder.Grow(16384)
	switch args[0] {
	case "bash":
		includeDesc := true
		if err := cmd.Root().GenBashCompletionV2(&builder, includeDesc); err != nil {
			return err
		}
	case "fish":
		includeDesc := true
		if err := cmd.Root().GenFishCompletion(&builder, includeDesc); err != nil {
			return err
		}
	case "powershell":
		if err := cmd.Root().GenPowerShellCompletion(&builder); err != nil {
			return err
		}
	case "zsh":
		if err := cmd.Root().GenZshCompletion(&builder); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unsupported shell", args[0])
	}
	return c.writeOutputString(builder.String())
}
