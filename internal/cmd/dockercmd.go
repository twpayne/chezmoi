package cmd

import (
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type dockerCmdConfig struct {
	Command string `json:"command" mapstructure:"command" yaml:"command"`
	exec    dockerExecCmdConfig
	run     dockerRunCmdConfig
}

type dockerExecCmdConfig struct {
	interactive    bool
	packageManager string
	shell          bool
}

type dockerRunCmdConfig struct {
	packageManager string
}

func (c *Config) newDockerCmd() *cobra.Command {
	commandCmd := &cobra.Command{
		GroupID: groupIDRemote,
		Use:     "docker",
		Short:   "Use your dotfiles in a Docker container",
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}

	commandExecCmd := &cobra.Command{
		Use:   "exec container-id [args...]",
		Short: "Install your dotfiles in an existing Docker container and execute a shell",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.makeRunEWithSourceState(c.runDockerExecCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	commandExecCmd.Flags().
		BoolVarP(&c.Docker.exec.interactive, "interactive", "i", c.Docker.exec.interactive, "Run interactively")
	commandExecCmd.Flags().
		StringVarP(&c.Docker.exec.packageManager, "package-manager", "p", c.Docker.exec.packageManager, "Package manager")
	commandExecCmd.Flags().BoolVarP(&c.Docker.exec.shell, "shell", "s", c.Docker.exec.shell, "Execute shell afterwards")
	commandCmd.AddCommand(commandExecCmd)

	commandRunCmd := &cobra.Command{
		Use:   "run image-id [args...]",
		Short: "Create a new Docker container with your dotfiles and execute a shell",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.makeRunEWithSourceState(c.runDockerRunCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	commandRunCmd.Flags().
		StringVarP(&c.Docker.run.packageManager, "package-manager", "p", c.Docker.run.packageManager, "Package manager")
	commandCmd.AddCommand(commandRunCmd)

	return commandCmd
}

func (c *Config) runDockerExecCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	commandArgs := []string{"exec"}
	if c.Docker.exec.interactive {
		commandArgs = append(commandArgs,
			"--interactive",
			"--tty",
		)
	}
	commandArgs = append(commandArgs, args[0])
	return c.runInstallInitShellSh(sourceState,
		c.Docker.Command, commandArgs,
		runInstallInitShellOptions{
			args:           args[1:],
			interactive:    c.Docker.exec.interactive,
			packageManager: c.Docker.exec.packageManager,
			shell:          c.Docker.exec.shell,
		},
	)
}

func (c *Config) runDockerRunCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	return c.runInstallInitShellSh(sourceState,
		c.Docker.Command, []string{"run", "--interactive", "--tty", args[0]},
		runInstallInitShellOptions{
			args:           args[1:],
			interactive:    true,
			packageManager: c.Docker.run.packageManager,
			shell:          true,
		},
	)
}
