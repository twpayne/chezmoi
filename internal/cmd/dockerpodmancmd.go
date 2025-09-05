package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type dockerPodmanCmdConfig struct {
	exec dockerPodmanExecCmdConfig
	run  dockerPodmanRunCmdConfig
}

type dockerPodmanExecCmdConfig struct {
	interactive    bool
	packageManager string
	shell          bool
}

type dockerPodmanRunCmdConfig struct {
	packageManager string
}

func (c *Config) newDockerPodmanCmd(command string, config *dockerPodmanCmdConfig) *cobra.Command {
	commandCmd := &cobra.Command{
		Use:   command,
		Short: "Install chezmoi and your dotfiles in a " + command + " container",
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}

	commandExecCmd := &cobra.Command{
		Use:   "exec container-id [args...]",
		Short: "Install chezmoi and your dotfiles in an existing " + command + " container",
		Args:  cobra.MinimumNArgs(1),
		RunE: c.makeRunEWithSourceState(func(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
			commandArgs := []string{"exec"}
			if config.exec.interactive {
				commandArgs = append(commandArgs,
					"--interactive",
					"--tty",
				)
			}
			commandArgs = append(commandArgs, args[0])
			return c.runInstallInitShellSh(sourceState,
				command, commandArgs,
				runInstallInitShellOptions{
					args:           args[1:],
					interactive:    config.exec.interactive,
					packageManager: config.exec.packageManager,
					shell:          config.exec.shell,
				},
			)
		}),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	commandExecCmd.Flags().
		BoolVarP(&config.exec.interactive, "interactive", "i", config.exec.interactive, "Run interactively")
	commandExecCmd.Flags().
		StringVarP(&config.exec.packageManager, "package-manager", "p", config.exec.packageManager, "Package manager")
	commandExecCmd.Flags().BoolVarP(&config.exec.shell, "shell", "s", config.exec.shell, "Execute shell afterwards")
	commandCmd.AddCommand(commandExecCmd)

	commandRunCmd := &cobra.Command{
		Use:   "run image-id [args...]",
		Short: "Install chezmoi and your dotfiles in a new " + command + " container",
		Args:  cobra.MinimumNArgs(1),
		RunE: c.makeRunEWithSourceState(func(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
			return c.runInstallInitShellSh(sourceState,
				command, []string{"run", "--interactive", "--tty", args[0]},
				runInstallInitShellOptions{
					args:           args[1:],
					interactive:    true,
					packageManager: config.run.packageManager,
					shell:          true,
				},
			)
		}),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	commandRunCmd.Flags().
		StringVarP(&config.run.packageManager, "package-manager", "p", config.run.packageManager, "Package manager")
	commandCmd.AddCommand(commandRunCmd)

	return commandCmd
}
