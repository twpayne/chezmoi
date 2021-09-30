package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type editCmdConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
	apply   bool
	exclude *chezmoi.EntryTypeSet
	include *chezmoi.EntryTypeSet
	init    bool
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
	flags.BoolVarP(&c.Edit.apply, "apply", "a", c.Edit.apply, "Apply after editing")
	flags.VarP(c.Edit.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Edit.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Edit.init, "init", c.update.init, "Recreate config file from template")

	return editCmd
}

func (c *Config) runEditCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	if len(args) == 0 {
		if err := c.runEditor([]string{c.WorkingTreeAbsPath.String()}); err != nil {
			return err
		}
		if c.Edit.apply {
			if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, noArgs, applyArgsOptions{
				include:      c.Edit.include.Sub(c.Edit.exclude),
				init:         c.Edit.init,
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
	type transparentlyDecryptedFile struct {
		sourceAbsPath    chezmoi.AbsPath
		decryptedAbsPath chezmoi.AbsPath
	}
	var transparentlyDecryptedFiles []transparentlyDecryptedFile
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		sourceRelPath := sourceStateEntry.SourceRelPath()
		if sourceStateFile, ok := sourceStateEntry.(*chezmoi.SourceStateFile); ok && sourceStateFile.Attr.Encrypted {
			tempDirAbsPath, err := c.tempDir("chezmoi-edit")
			if err != nil {
				return err
			}
			// FIXME use RawContents and DecryptFile
			decryptedAbsPath := tempDirAbsPath.Join(sourceRelPath.TargetRelPath(c.encryption.EncryptedSuffix()))
			contents, err := sourceStateFile.Contents()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(decryptedAbsPath.Dir().String(), 0o700); err != nil {
				return err
			}
			if err := c.baseSystem.WriteFile(decryptedAbsPath, contents, 0o600); err != nil {
				return err
			}
			transparentlyDecryptedFile := transparentlyDecryptedFile{
				sourceAbsPath:    c.SourceDirAbsPath.Join(sourceRelPath.RelPath()),
				decryptedAbsPath: decryptedAbsPath,
			}
			transparentlyDecryptedFiles = append(transparentlyDecryptedFiles, transparentlyDecryptedFile)
			editorArgs = append(editorArgs, decryptedAbsPath.String())
		} else {
			sourceAbsPath := c.SourceDirAbsPath.Join(sourceRelPath.RelPath())
			editorArgs = append(editorArgs, sourceAbsPath.String())
		}
	}

	if err := c.runEditor(editorArgs); err != nil {
		return err
	}

	for _, transparentlyDecryptedFile := range transparentlyDecryptedFiles {
		contents, err := c.encryption.EncryptFile(transparentlyDecryptedFile.decryptedAbsPath)
		if err != nil {
			return err
		}
		if err := c.baseSystem.WriteFile(transparentlyDecryptedFile.sourceAbsPath, contents, 0o666); err != nil {
			return err
		}
	}

	if c.Edit.apply {
		if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
			include:      c.Edit.include,
			init:         c.Edit.init,
			recursive:    false,
			umask:        c.Umask,
			preApplyFunc: c.defaultPreApplyFunc,
		}); err != nil {
			return err
		}
	}

	return nil
}
