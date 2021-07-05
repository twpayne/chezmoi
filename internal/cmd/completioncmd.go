package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

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
	var sb strings.Builder
	sb.Grow(16384)
	switch args[0] {
	case "bash":
		includeDesc := true
		if err := cmd.Root().GenBashCompletionV2(&sb, includeDesc); err != nil {
			return err
		}
	case "fish":
		includeDesc := true
		if err := cmd.Root().GenFishCompletion(&sb, includeDesc); err != nil {
			return err
		}
	case "powershell":
		if err := cmd.Root().GenPowerShellCompletion(&sb); err != nil {
			return err
		}
	case "zsh":
		if err := cmd.Root().GenZshCompletion(&sb); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unsupported shell", args[0])
	}
	return c.writeOutputString(sb.String())
}
