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
	var output string
	switch args[0] {
	case "bash":
		includeDesc := true
		if err := cmd.Root().GenBashCompletionV2(&builder, includeDesc); err != nil {
			return err
		}
		output = builder.String()
	case "fish":
		includeDesc := true
		if err := cmd.Root().GenFishCompletion(&builder, includeDesc); err != nil {
			return err
		}
		output = builder.String()
	case "powershell":
		if err := cmd.Root().GenPowerShellCompletion(&builder); err != nil {
			return err
		}
		output = builder.String()
	case "zsh":
		if err := cmd.Root().GenZshCompletion(&builder); err != nil {
			return err
		}
		// FIXME remove the following when
		// https://github.com/spf13/cobra/pull/1690 is merged and released
		output = "#compdef chezmoi\n" + strings.TrimPrefix(builder.String(), "#compdef _chezmoi chezmoi\n")

	default:
		return fmt.Errorf("%s: unsupported shell", args[0])
	}
	return c.writeOutputString(output)
}
