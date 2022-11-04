package cmd

import (
	"os"
	"runtime"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type editCmdConfig struct {
	Command     string        `json:"command" mapstructure:"command" yaml:"command"`
	Args        []string      `json:"args" mapstructure:"args" yaml:"args"`
	Hardlink    bool          `json:"hardlink" mapstructure:"hardlink" yaml:"hardlink"`
	MinDuration time.Duration `json:"minDuration" mapstructure:"minDuration" yaml:"minDuration"`
	Watch       bool          `json:"watch" mapstructure:"watch" yaml:"watch"`
	Apply       bool          `json:"apply" mapstructure:"apply" yaml:"apply"`
	filter      *chezmoi.EntryTypeFilter
	init        bool
}

func (c *Config) newEditCmd() *cobra.Command {
	editCmd := &cobra.Command{
		Use:               "edit targets...",
		Short:             "Edit the source state of a target",
		Long:              mustLongHelp("edit"),
		Example:           example("edit"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runEditCmd,
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			modifiesSourceDirectory:      "true",
			persistentStateMode:          persistentStateModeReadWrite,
			requiresSourceDirectory:      "true",
			runsCommands:                 "true",
		},
	}

	flags := editCmd.Flags()
	flags.BoolVarP(&c.Edit.Apply, "apply", "a", c.Edit.Apply, "Apply after editing")
	flags.VarP(c.Edit.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.BoolVar(&c.Edit.Hardlink, "hardlink", c.Edit.Hardlink, "Invoke editor with a hardlink to the source file")
	flags.VarP(c.Edit.filter.Include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Edit.init, "init", c.Edit.init, "Recreate config file from template")
	flags.BoolVar(&c.Edit.Watch, "watch", c.Edit.Watch, "Apply on save")

	registerExcludeIncludeFlagCompletionFuncs(editCmd)

	return editCmd
}

func (c *Config) runEditCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if err := c.runEditor([]string{c.WorkingTreeAbsPath.String()}); err != nil {
			return err
		}
		if c.Edit.Apply {
			if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, noArgs, applyArgsOptions{
				filter:       c.Edit.filter,
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

	sourceState, err := c.newSourceState(cmd.Context())
	if err != nil {
		return err
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
TARGETRELPATH:
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		sourceRelPath := sourceStateEntry.SourceRelPath()
		switch sourceStateFile, ok := sourceStateEntry.(*chezmoi.SourceStateFile); {
		case ok && sourceStateFile.Attr.Encrypted:
			// FIXME in the case that the file is an encrypted template then we
			// should first decrypt the file to a temporary directory and
			// secondly add a hardlink from the edit directory to the temporary
			// directory.

			tempDirAbsPath, err := c.tempDir("chezmoi-encrypted")
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
		case ok && c.Edit.Hardlink && runtime.GOOS != "windows":
			// If the operating system supports hard links and the file is not
			// encrypted, then create a hard link to the file in the source
			// directory in the temporary edit directory. This means that the
			// editor will see the target filename while simultaneously updating
			// the file in the source directory.

			// Compute the hard link path from the target path. If the file is a
			// template then preserve the .tmpl suffix as a clue to the editor.
			targetRelPath := sourceRelPath.TargetRelPath(c.encryption.EncryptedSuffix())
			if sourceStateFile.Attr.Template {
				targetRelPath = targetRelPath.AppendString(chezmoi.TemplateSuffix)
			}
			tempDirAbsPath, err := c.tempDir("chezmoi-edit")
			if err != nil {
				return err
			}
			hardlinkAbsPath := tempDirAbsPath.Join(targetRelPath)

			// Attempt to create the hard link. If this succeeds, continue to
			// the next target. Hardlinking will fail if the temporary directory
			// is on a different filesystem to the source directory, which is
			// not the case for most users.
			//
			// FIXME create a temporary directory on the same filesystem as the
			// source directory if needed.
			if err := os.MkdirAll(hardlinkAbsPath.Dir().String(), 0o700); err != nil {
				return err
			}
			if err := c.baseSystem.Link(c.SourceDirAbsPath.Join(sourceRelPath.RelPath()), hardlinkAbsPath); err == nil {
				editorArgs = append(editorArgs, hardlinkAbsPath.String())
				continue TARGETRELPATH
			}

			// Otherwise, fall through to the default option of editing the
			// source file in the source state.
			fallthrough
		default:
			sourceAbsPath := c.SourceDirAbsPath.Join(sourceRelPath.RelPath())
			editorArgs = append(editorArgs, sourceAbsPath.String())
		}
	}

	postEditFunc := func() error {
		for _, transparentlyDecryptedFile := range transparentlyDecryptedFiles {
			contents, err := c.encryption.EncryptFile(transparentlyDecryptedFile.decryptedAbsPath)
			if err != nil {
				return err
			}
			if err := c.baseSystem.WriteFile(transparentlyDecryptedFile.sourceAbsPath, contents, 0o666); err != nil {
				return err
			}
		}

		if c.Edit.Apply || c.Edit.Watch {
			if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
				filter:       c.Edit.filter,
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

	if c.Edit.Watch {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer watcher.Close()

		for _, editorArg := range editorArgs {
			// FIXME watch directories recursively
			if err := watcher.Add(editorArg); err != nil {
				return err
			}
		}

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					c.logger.Debug().
						Stringer("Op", event.Op).
						Str("Name", event.Name).
						Msg("watcher.Events")
					err := postEditFunc()
					c.logger.Err(err).
						Msg("postEditFunc")
				case _, ok := <-watcher.Errors:
					if !ok {
						return
					}
					c.logger.Error().
						Err(err).
						Msg("watcher.Errors")
				}
			}
		}()
	}

	if err := c.runEditor(editorArgs); err != nil {
		return err
	}

	if err := postEditFunc(); err != nil {
		return err
	}

	return nil
}
