package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/go-shell"
	"github.com/twpayne/go-vfs"
	bolt "go.etcd.io/bbolt"
)

type diffCmdConfig struct {
	Format  string
	NoPager bool
	Pager   string
}

var diffCmd = &cobra.Command{
	Use:     "diff [targets...]",
	Short:   "Print the diff between the target state and the destination state",
	Long:    mustGetLongHelp("diff"),
	Example: getExample("diff"),
	PreRunE: config.ensureNoError,
	RunE:    config.runDiffCmd,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	persistentFlags := diffCmd.PersistentFlags()
	persistentFlags.StringVarP(&config.Diff.Format, "format", "f", config.Diff.Format, "format, \"chezmoi\" or \"git\"")
	persistentFlags.BoolVar(&config.Diff.NoPager, "no-pager", false, "disable pager")

	markRemainingZshCompPositionalArgumentsAsFiles(diffCmd, 1)
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	c.DryRun = true // Prevent scripts from running.

	switch c.Diff.Format {
	case "chezmoi":
		c.mutator = chezmoi.NullMutator{}
	case "git":
		c.mutator = chezmoi.NewFSMutator(vfs.NewReadOnlyFS(config.fs))
	default:
		return fmt.Errorf("unknown diff format: %q", c.Diff.Format)
	}
	if c.Debug {
		c.mutator = chezmoi.NewDebugMutator(c.mutator)
	}

	persistentState, err := c.getPersistentState(&bolt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}
	defer persistentState.Close()

	if c.Diff.NoPager || c.Diff.Pager == "" {
		switch c.Diff.Format {
		case "chezmoi":
			c.mutator = chezmoi.NewVerboseMutator(c.Stdout, c.mutator, c.colored, c.maxDiffDataSize)
		case "git":
			unifiedEncoder := diff.NewUnifiedEncoder(c.Stdout, diff.DefaultContextLines)
			if c.colored {
				unifiedEncoder.SetColor(diff.NewColorConfig())
			}
			c.mutator = chezmoi.NewGitDiffMutator(unifiedEncoder, c.mutator, c.DestDir+string(filepath.Separator))
		}
		return c.applyArgs(args, persistentState)
	}

	var pagerCmd *exec.Cmd
	var pagerStdinPipe io.WriteCloser

	// If the pager command contains any spaces, assume that it is a full
	// shell command to be executed via the user's shell. Otherwise, execute
	// it directly.
	if strings.IndexFunc(c.Diff.Pager, unicode.IsSpace) != -1 {
		shell, _ := shell.CurrentUserShell()
		//nolint:gosec
		pagerCmd = exec.Command(shell, "-c", c.Diff.Pager)
	} else {
		//nolint:gosec
		pagerCmd = exec.Command(c.Diff.Pager)
	}
	pagerStdinPipe, err = pagerCmd.StdinPipe()
	if err != nil {
		return err
	}
	pagerCmd.Stdout = c.Stdout
	pagerCmd.Stderr = c.Stderr
	if err := pagerCmd.Start(); err != nil {
		return err
	}

	switch c.Diff.Format {
	case "chezmoi":
		c.mutator = chezmoi.NewVerboseMutator(pagerStdinPipe, c.mutator, c.colored, c.maxDiffDataSize)
	case "git":
		unifiedEncoder := diff.NewUnifiedEncoder(pagerStdinPipe, diff.DefaultContextLines)
		if c.colored {
			unifiedEncoder.SetColor(diff.NewColorConfig())
		}
		c.mutator = chezmoi.NewGitDiffMutator(unifiedEncoder, c.mutator, c.DestDir+string(filepath.Separator))
	}

	if err := c.applyArgs(args, persistentState); err != nil {
		return err
	}

	if err := pagerStdinPipe.Close(); err != nil {
		return err
	}

	return pagerCmd.Wait()
}
