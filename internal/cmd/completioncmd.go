package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type completionCmdConfig struct {
	Custom bool `json:"custom" mapstructure:"custom" yaml:"custom"`
}

func (c *Config) newCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		GroupID:   groupIDInternal,
		Use:       "completion shell",
		Short:     "Generate shell completion code",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "fish", "powershell", "zsh"},
		Long:      mustLongHelp("completion"),
		Example:   example("completion"),
		RunE:      c.runCompletionCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}

	return completionCmd
}

func (c *Config) runCompletionCmd(cmd *cobra.Command, args []string) error {
	completion, err := completion(cmd, args[0])
	if err != nil {
		return err
	}
	return c.writeOutputString(completion, 0o666)
}

func completion(cmd *cobra.Command, shell string) (string, error) {
	builder := strings.Builder{}
	builder.Grow(16384)
	switch shell {
	case "bash":
		includeDesc := true
		if err := cmd.Root().GenBashCompletionV2(&builder, includeDesc); err != nil {
			return "", err
		}
	case "fish":
		includeDesc := true
		if err := cmd.Root().GenFishCompletion(&builder, includeDesc); err != nil {
			return "", err
		}
	case "powershell":
		if err := cmd.Root().GenPowerShellCompletionWithDesc(&builder); err != nil {
			return "", err
		}
	case "zsh":
		if err := cmd.Root().GenZshCompletion(&builder); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("%s: unsupported shell", shell)
	}
	return builder.String(), nil
}
