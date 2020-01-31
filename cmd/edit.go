package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/renameio"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var editCmd = &cobra.Command{
	Use:      "edit targets...",
	Short:    "Edit the source state of a target",
	Long:     mustGetLongHelp("edit"),
	Example:  getExample("edit"),
	PreRunE:  config.ensureNoError,
	RunE:     config.runEditCmd,
	PostRunE: config.autoCommitAndAutoPush,
}

type editCmdConfig struct {
	apply  bool
	diff   bool
	prompt bool
}

func init() {
	rootCmd.AddCommand(editCmd)

	persistentFlags := editCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.edit.apply, "apply", "a", false, "apply edit after editing")
	persistentFlags.BoolVarP(&config.edit.diff, "diff", "d", false, "print diff after editing")
	persistentFlags.BoolVarP(&config.edit.prompt, "prompt", "p", false, "prompt before applying (implies --diff)")

	markRemainingZshCompPositionalArgumentsAsFiles(editCmd, 1)
}

type encryptedFile struct {
	index          int
	file           *chezmoi.File
	ciphertextPath string
	plaintextPath  string
}

func (c *Config) runEditCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if c.edit.apply {
			c.warn("--apply is currently ignored when edit is run with no arguments")
		}
		if c.edit.diff {
			c.warn("--diff is currently ignored when edit is run with no arguments")
		}
		if c.edit.prompt {
			c.warn("--prompt is currently ignored when edit is run with no arguments")
		}
		return c.run("", c.getEditor(), c.SourceDir)
	}

	if c.edit.prompt {
		c.edit.diff = true
	}

	ts, err := c.getTargetState(&chezmoi.PopulateOptions{
		ExecuteTemplates: false,
	})
	if err != nil {
		return err
	}

	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}

	// Build a list of source file names to pass to the editor. Check that each
	// is either a file or a symlink. If the entry is an encrypted file then
	// remember it.
	argv := make([]string, len(entries))
	var encryptedFiles []encryptedFile
	for i, entry := range entries {
		argv[i] = filepath.Join(c.SourceDir, entry.SourceName())
		if file, ok := entry.(*chezmoi.File); ok {
			if file.Encrypted {
				ef := encryptedFile{
					index:          i,
					file:           file,
					ciphertextPath: argv[i],
				}
				encryptedFiles = append(encryptedFiles, ef)
			}
		} else if _, ok := entry.(*chezmoi.Symlink); !ok {
			return fmt.Errorf("%s: not a file or symlink", args[i])
		}
	}

	// If any of the files are encrypted, create a temporary directory to store
	// the plaintext contents, decrypt each of them, and update argv to point to
	// the plaintext file.
	if len(encryptedFiles) != 0 {
		tempDir, err := ioutil.TempDir("", "chezmoi")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		for i := range encryptedFiles {
			ef := &encryptedFiles[i]
			plaintext, err := ef.file.Contents()
			if err != nil {
				return err
			}
			ef.plaintextPath = filepath.Join(tempDir, ef.file.SourceName())
			if err := os.MkdirAll(filepath.Dir(ef.plaintextPath), 0700); err != nil {
				return err
			}
			if err := renameio.WriteFile(ef.plaintextPath, plaintext, 0600); err != nil {
				return err
			}
			argv[ef.index] = ef.plaintextPath
		}
	}

	if err := c.runEditor(argv...); err != nil {
		return err
	}

	// Re-encrypt any encrypted files.
	for _, ef := range encryptedFiles {
		plaintext, err := ioutil.ReadFile(ef.plaintextPath)
		if err != nil {
			return err
		}
		ciphertext, err := ts.GPG.Encrypt(ef.plaintextPath, plaintext)
		if err != nil {
			return err
		}
		if err := renameio.WriteFile(ef.ciphertextPath, ciphertext, 0644); err != nil {
			return err
		}
	}

	// Recompute the target state and entries after editing.
	ts, err = c.getTargetState(nil)
	if err != nil {
		return err
	}

	entries, err = c.getEntries(ts, args)
	if err != nil {
		return err
	}

	readOnlyFS := vfs.NewReadOnlyFS(c.fs)
	applyOptions := chezmoi.ApplyOptions{
		DestDir:           ts.DestDir,
		DryRun:            c.DryRun,
		Ignore:            ts.TargetIgnore.Match,
		ScriptStateBucket: c.scriptStateBucket,
		Stdout:            c.Stdout(),
		Umask:             ts.Umask,
		Verbose:           c.Verbose,
	}
	for i, entry := range entries {
		anyMutator := chezmoi.NewAnyMutator(chezmoi.NullMutator{})
		var mutator chezmoi.Mutator = anyMutator
		if c.edit.diff {
			mutator = chezmoi.NewVerboseMutator(c.Stdout(), mutator, c.colored, c.maxDiffDataSize)
		}
		if err := entry.Apply(readOnlyFS, mutator, c.Follow, &applyOptions); err != nil {
			return err
		}
		if c.edit.apply && anyMutator.Mutated() {
			if c.edit.prompt {
				choice, err := c.prompt(fmt.Sprintf("Apply %s", args[i]), "ynqa")
				if err != nil {
					return err
				}
				switch choice {
				case 'y':
				case 'n':
					continue
				case 'q':
					return nil
				case 'a':
					c.edit.prompt = false
				}
			}
			if err := entry.Apply(readOnlyFS, c.mutator, c.Follow, &applyOptions); err != nil {
				return err
			}
		}
	}
	return nil
}
