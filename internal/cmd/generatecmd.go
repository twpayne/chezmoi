package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/assets/templates"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/chezmoigit"
)

type generateCmdConfig struct {
	installInitShellSh generateInstallInitShellShCmdConfig
}

type generateInstallInitShellShCmdConfig struct {
	interactive bool
	_package    bool
	shell       bool
}

func (c *Config) newGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:     "generate file",
		Short:   "Generate a file for use with chezmoi",
		Long:    mustLongHelp("generate"),
		Example: example("generate"),
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}

	generateGitCommitMessageCmd := &cobra.Command{
		Use:   "git-commit-message",
		Short: "Generate a git commit message",
		Args:  cobra.NoArgs,
		RunE:  c.runGenerateGitCommitMessageCmd,
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}
	generateCmd.AddCommand(generateGitCommitMessageCmd)

	generateInstallShCmd := &cobra.Command{
		Use:   "install.sh",
		Short: "Generate an install script",
		Args:  cobra.NoArgs,
		RunE:  c.runGenerateInstallShCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	generateCmd.AddCommand(generateInstallShCmd)

	generateInstallInitShellShCmd := &cobra.Command{
		Use:   "install-init-shell.sh",
		Short: "Generate an install script that also executes a shell",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.makeRunEWithSourceState(c.runGenerateInstallInitShellShCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}
	generateInstallInitShellShCmd.Flags().
		BoolVarP(&c.generate.installInitShellSh.interactive, "interactive", "i", c.generate.installInitShellSh.interactive, "Set interactive")
	generateInstallInitShellShCmd.Flags().
		BoolVarP(&c.generate.installInitShellSh.shell, "package", "p", c.generate.installInitShellSh._package, "Install with package")
	generateInstallInitShellShCmd.Flags().
		BoolVarP(&c.generate.installInitShellSh.shell, "shell", "s", c.generate.installInitShellSh.shell, "Set shell")
	generateCmd.AddCommand(generateInstallInitShellShCmd)

	return generateCmd
}

func (c *Config) runGenerateGitCommitMessageCmd(cmd *cobra.Command, args []string) error {
	builder := strings.Builder{}
	builder.Grow(16384)
	output, err := c.cmdOutput(c.WorkingTreeAbsPath, c.Git.Command, []string{"status", "--porcelain=v2"})
	if err != nil {
		return err
	}
	status, err := chezmoigit.ParseStatusPorcelainV2(output)
	if err != nil {
		return err
	}
	data, err := c.gitCommitMessage(cmd, status)
	if err != nil {
		return err
	}
	if _, err := builder.Write(data); err != nil {
		return err
	}
	return c.writeOutputString(builder.String(), 0o666)
}

func (c *Config) runGenerateInstallShCmd(cmd *cobra.Command, args []string) error {
	return c.writeOutput(templates.InstallSh, 0o777)
}

func (c *Config) runGenerateInstallInitShellShCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	script, err := sourceState.ExecuteTemplateData(chezmoi.ExecuteTemplateDataOptions{
		Name: "install-init-shell.sh.tmpl",
		Data: templates.InstallInitShellShTmpl,
		ExtraData: map[string]any{
			"args":        args,
			"interactive": c.generate.installInitShellSh.interactive,
			"package":     c.generate.installInitShellSh._package,
			"shell":       c.generate.installInitShellSh.shell,
		},
	})
	if err != nil {
		return err
	}
	return c.writeOutput(script, 0o777)
}
