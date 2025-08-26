package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type sshCmdConfig struct {
	_package bool
	shell    bool
}

func (c *Config) newSSHCmd() *cobra.Command {
	sshCmd := &cobra.Command{
		Use:     "ssh host",
		Short:   "ssh to a host and initialize dotfiles",
		Long:    mustLongHelp("ssh"),
		Example: example("ssh"),
		Args:    cobra.MinimumNArgs(1),
		RunE:    c.makeRunEWithSourceState(c.runSSHCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	sshCmd.Flags().BoolVarP(&c.ssh._package, "package", "p", c.ssh._package, "Install with package")
	sshCmd.Flags().BoolVarP(&c.ssh.shell, "shell", "s", c.ssh.shell, "Execute shell afterwards")

	return sshCmd
}

func (c *Config) runSSHCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	return c.runInstallInitShellSh(sourceState,
		"ssh", []string{args[0]},
		runInstallInitShellOptions{
			args:        args[1:],
			interactive: true,
			_package:    c.ssh._package,
			shell:       c.ssh.shell,
		},
	)
}
