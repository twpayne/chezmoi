package cmd

import (
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type sshCmdConfig struct {
	packageManager string
	shell          bool
}

func (c *Config) newSSHCmd() *cobra.Command {
	sshCmd := &cobra.Command{
		Use:     "ssh host",
		Short:   "SSH to a host and initialize dotfiles",
		Long:    mustLongHelp("ssh"),
		Example: example("ssh"),
		Args:    cobra.MinimumNArgs(1),
		RunE:    c.makeRunEWithSourceState(c.runSSHCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	sshCmd.Flags().StringVarP(&c.ssh.packageManager, "package-manager", "p", c.ssh.packageManager, "Package manager")
	sshCmd.Flags().BoolVarP(&c.ssh.shell, "shell", "s", c.ssh.shell, "Execute shell afterwards")

	return sshCmd
}

func (c *Config) runSSHCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	return c.runInstallInitShellSh(sourceState,
		"ssh", []string{args[0]},
		runInstallInitShellOptions{
			args:           args[1:],
			interactive:    true,
			packageManager: c.ssh.packageManager,
			shell:          c.ssh.shell,
		},
	)
}
