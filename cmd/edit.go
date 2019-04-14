package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/renameio"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var editCmd = &cobra.Command{
	Use:     "edit targets...",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Edit the source state of a target",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runEditCmd),
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
}

type encryptedFile struct {
	index          int
	file           *chezmoi.File
	ciphertextPath string
	plaintextPath  string
}

func (c *Config) runEditCmd(fs vfs.FS, args []string) error {
	if c.edit.prompt {
		c.edit.diff = true
	}

	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}

	entries, err := c.getEntries(fs, ts, args)
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

	// Short path: if no post-edit actions are required then exec the editor
	// directly.
	if !c.edit.diff && !c.edit.apply && len(encryptedFiles) == 0 {
		return c.execEditor(argv...)
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

	// Re-encrypt any encypted files.
	for _, ef := range encryptedFiles {
		plaintext, err := ioutil.ReadFile(ef.plaintextPath)
		if err != nil {
			return err
		}
		ciphertext, err := ts.Encrypt(plaintext)
		if err != nil {
			return err
		}
		if err := renameio.WriteFile(ef.ciphertextPath, ciphertext, 0644); err != nil {
			return err
		}
	}

	readOnlyFS := vfs.NewReadOnlyFS(fs)
	applyMutator := c.getDefaultMutator(fs)
	for i, entry := range entries {
		anyMutator := chezmoi.NewAnyMutator(chezmoi.NullMutator)
		var mutator chezmoi.Mutator = anyMutator
		if c.edit.diff {
			mutator = chezmoi.NewLoggingMutator(c.Stdout(), mutator)
		}
		if err := entry.Apply(readOnlyFS, ts.DestDir, ts.TargetIgnore.Match, ts.Umask, mutator); err != nil {
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
			if err := entry.Apply(readOnlyFS, ts.DestDir, ts.TargetIgnore.Match, ts.Umask, applyMutator); err != nil {
				return err
			}
		}
	}
	return nil
}
