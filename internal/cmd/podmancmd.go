package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type podmanCmdConfig struct {
	exec podmanExecCmdConfig
	run  podmanRunCmdConfig
}

type podmanExecCmdConfig struct {
	interactive bool
	_package    bool
	shell       bool
}

type podmanRunCmdConfig struct {
	_package bool
}

func (c *Config) newPodmanCmd() *cobra.Command {
	podmanCmd := &cobra.Command{
		Use:   "podman",
		Short: "Install chezmoi and your dotfiles in a podman container",
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}

	podmanExecCmd := &cobra.Command{
		Use:   "exec container-id [args...]",
		Short: "Install chezmoi and your dotfiles in an existing podman container",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.makeRunEWithSourceState(c.runPodmanExecCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	podmanExecCmd.Flags().
		BoolVarP(&c.podman.exec.interactive, "interactive", "i", c.podman.exec.interactive, "Run interactively")
	podmanExecCmd.Flags().BoolVarP(&c.podman.exec._package, "package", "p", c.podman.exec._package, "Install with package")
	podmanExecCmd.Flags().BoolVarP(&c.podman.exec.shell, "shell", "s", c.podman.exec.shell, "Execute shell afterwards")
	podmanCmd.AddCommand(podmanExecCmd)

	podmanRunCmd := &cobra.Command{
		Use:   "run image-id [args...]",
		Short: "Install chezmoi and your dotfiles in a new podman container",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.makeRunEWithSourceState(c.runPodmanRun),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	podmanRunCmd.Flags().BoolVarP(&c.podman.run._package, "package", "p", c.podman.run._package, "Install with package")
	podmanCmd.AddCommand(podmanRunCmd)

	return podmanCmd
}

func (c *Config) runPodmanExecCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	podmanArgs := []string{"exec"}
	if c.podman.exec.interactive {
		podmanArgs = append(podmanArgs,
			"--interactive",
			"--tty",
		)
	}
	podmanArgs = append(podmanArgs, args[0])
	return c.runInstallInitShellSh(sourceState,
		"podman", podmanArgs,
		runInstallInitShellOptions{
			args:        args[1:],
			interactive: c.podman.exec.interactive,
			_package:    c.podman.exec._package,
			shell:       c.podman.exec.shell,
		},
	)
}

func (c *Config) runPodmanRun(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	return c.runInstallInitShellSh(sourceState,
		"podman", []string{"run", "--interactive", "--tty", args[0]},
		runInstallInitShellOptions{
			args:        args[1:],
			interactive: true,
			_package:    c.podman.run._package,
			shell:       true,
		},
	)
}
