package cmd

import (
	"io/ioutil"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type editCmdConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
	apply   bool
	include *chezmoi.IncludeSet
}

func (c *Config) newEditCmd() *cobra.Command {
	editCmd := &cobra.Command{
		Use:     "edit targets...",
		Short:   "Edit the source state of a target",
		Long:    mustLongHelp("edit"),
		Example: example("edit"),
		RunE:    c.makeRunEWithSourceState(c.runEditCmd),
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			modifiesSourceDirectory:      "true",
			persistentStateMode:          persistentStateModeReadWrite,
			requiresSourceDirectory:      "true",
			runsCommands:                 "true",
		},
	}

	flags := editCmd.Flags()
	flags.BoolVarP(&c.Edit.apply, "apply", "a", c.Edit.apply, "apply edit after editing")
	flags.VarP(c.Edit.include, "include", "i", "include entry types")

	return editCmd
}

func (c *Config) runEditCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	if len(args) == 0 {
		if err := c.runEditor([]string{string(c.sourceDirAbsPath)}); err != nil {
			return err
		}
		if c.Edit.apply {
			if err := c.applyArgs(c.destSystem, c.destDirAbsPath, noArgs, applyArgsOptions{
				include:      c.Edit.include,
				recursive:    true,
				umask:        c.Umask,
				preApplyFunc: c.defaultPreApplyFunc,
			}); err != nil {
				return err
			}
		}
		return nil
	}

	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	editorArgs := make([]string, 0, len(targetRelPaths))
	var decryptedDirAbsPath chezmoi.AbsPath
	type transparentlyDecryptedFile struct {
		sourceAbsPath    chezmoi.AbsPath
		decryptedAbsPath chezmoi.AbsPath
	}
	var transparentlyDecryptedFiles []transparentlyDecryptedFile
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		sourceRelPath := sourceStateEntry.SourceRelPath().RelPath()
		var editorArg string
		if sourceStateFile, ok := sourceStateEntry.(*chezmoi.SourceStateFile); ok && sourceStateFile.Attr.Encrypted {
			if decryptedDirAbsPath == "" {
				decryptedDir, err := ioutil.TempDir("", "chezmoi-decrypted")
				if err != nil {
					return err
				}
				decryptedDirAbsPath = chezmoi.AbsPath(decryptedDir)
				defer func() {
					_ = c.baseSystem.RemoveAll(decryptedDirAbsPath)
				}()
				if runtime.GOOS != "windows" {
					if err := c.baseSystem.Chmod(decryptedDirAbsPath, 0o700); err != nil {
						return err
					}
				}
			}
			// FIXME use RawContents and DecryptFile
			decryptedAbsPath := decryptedDirAbsPath.Join(sourceRelPath)
			contents, err := sourceStateFile.Contents()
			if err != nil {
				return err
			}
			if err := c.baseSystem.WriteFile(decryptedAbsPath, contents, 0o600); err != nil {
				return err
			}
			transparentlyDecryptedFile := transparentlyDecryptedFile{
				sourceAbsPath:    c.sourceDirAbsPath.Join(sourceRelPath),
				decryptedAbsPath: decryptedAbsPath,
			}
			transparentlyDecryptedFiles = append(transparentlyDecryptedFiles, transparentlyDecryptedFile)
			editorArg = string(decryptedAbsPath)
		} else {
			sourceAbsPath := c.sourceDirAbsPath.Join(sourceRelPath)
			editorArg = string(sourceAbsPath)
		}
		editorArgs = append(editorArgs, editorArg)
	}

	if err := c.runEditor(editorArgs); err != nil {
		return err
	}

	for _, transparentlyDecryptedFile := range transparentlyDecryptedFiles {
		contents, err := c.encryption.EncryptFile(string(transparentlyDecryptedFile.decryptedAbsPath))
		if err != nil {
			return err
		}
		if err := c.baseSystem.WriteFile(transparentlyDecryptedFile.sourceAbsPath, contents, 0o666); err != nil {
			return err
		}
	}

	if c.Edit.apply {
		if err := c.applyArgs(c.destSystem, c.destDirAbsPath, args, applyArgsOptions{
			include:      c.Edit.include,
			recursive:    false,
			umask:        c.Umask,
			preApplyFunc: c.defaultPreApplyFunc,
		}); err != nil {
			return err
		}
	}

	return nil
}
