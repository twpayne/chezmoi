package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

func (c *Config) newEditEncryptedCmd() *cobra.Command {
	editEncryptedCmd := &cobra.Command{
		GroupID: groupIDEncryption,
		Use:     "edit-encrypted filename...",
		Short:   "Edit an encrypted file",
		Long:    mustLongHelp("edit-encrypted"),
		Example: example("edit-encrypted"),
		Args:    cobra.MinimumNArgs(1),
		RunE:    c.runEditEncryptedCmd,
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			persistentStateModeEmpty,
			runsCommands,
		),
	}
	return editEncryptedCmd
}

func (c *Config) runEditEncryptedCmd(cmd *cobra.Command, args []string) error {
	type argument struct {
		ciphertextAbsPath chezmoi.AbsPath
		plaintextAbsPath  chezmoi.AbsPath
	}

	arguments := make([]argument, len(args))

	// Write plaintexts to a temporary directory.
	tempDirAbsPath, err := c.tempDir("chezmoi-edit-encrypted")
	if err != nil {
		return err
	}
	for i, arg := range args {
		arg = filepath.Clean(arg)
		ciphertextAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return err
		}
		arguments[i].ciphertextAbsPath = ciphertextAbsPath
		ciphertext, err := c.baseSystem.ReadFile(ciphertextAbsPath)
		if err != nil {
			return err
		}
		relPath := ciphertextAbsPath.MustTrimDirPrefix(c.homeDirAbsPath)
		plaintextAbsPath := tempDirAbsPath.Join(relPath)
		if err := chezmoi.MkdirAll(c.baseSystem, plaintextAbsPath.Dir(), 0o700); err != nil {
			return err
		}
		if err := c.encryption.DecryptToFile(plaintextAbsPath, ciphertext); err != nil {
			return err
		}
		arguments[i].plaintextAbsPath = plaintextAbsPath
	}

	// Run the editor on the plaintexts.
	editorArgs := make([]string, len(arguments))
	for i, argument := range arguments {
		editorArgs[i] = argument.plaintextAbsPath.String()
	}
	if err := c.runEditor(editorArgs); err != nil {
		return err
	}

	// Write the ciphertexts.
	//
	// FIXME only write the ciphertext if the plaintext has changed
	// FIXME preserve original plaintext file mode
	for _, argument := range arguments {
		ciphertext, err := c.encryption.EncryptFile(argument.plaintextAbsPath)
		if err != nil {
			return err
		}
		if err := c.baseSystem.WriteFile(argument.ciphertextAbsPath, ciphertext, 0o666&^c.Umask); err != nil {
			return err
		}
	}

	return nil
}
