package cmd

import (
	"bytes"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
	"chezmoi.io/chezmoi/internal/chezmoilog"
)

type editCmdConfig struct {
	Command     string        `json:"command"     mapstructure:"command"     yaml:"command"`
	Args        []string      `json:"args"        mapstructure:"args"        yaml:"args"`
	Hardlink    bool          `json:"hardlink"    mapstructure:"hardlink"    yaml:"hardlink"`
	MinDuration time.Duration `json:"minDuration" mapstructure:"minDuration" yaml:"minDuration"`
	Watch       bool          `json:"watch"       mapstructure:"watch"       yaml:"watch"`
	Apply       bool          `json:"apply"       mapstructure:"apply"       yaml:"apply"`
	filter      *chezmoi.EntryTypeFilter
	init        bool
}

func (c *Config) newEditCmd() *cobra.Command {
	editCmd := &cobra.Command{
		GroupID:           groupIDDaily,
		Use:               "edit targets...",
		Short:             "Edit the source state of a target",
		Long:              mustLongHelp("edit"),
		Example:           example("edit"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runEditCmd,
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
			requiresSourceDirectory,
			runsCommands,
		),
	}

	editCmd.Flags().BoolVarP(&c.Edit.Apply, "apply", "a", c.Edit.Apply, "Apply after editing")
	editCmd.Flags().VarP(c.Edit.filter.Exclude, "exclude", "x", "Exclude entry types")
	editCmd.Flags().BoolVar(&c.Edit.Hardlink, "hardlink", c.Edit.Hardlink, "Invoke editor with a hardlink to the source file")
	editCmd.Flags().VarP(c.Edit.filter.Include, "include", "i", "Include entry types")
	editCmd.Flags().BoolVar(&c.Edit.init, "init", c.Edit.init, "Recreate config file from template")
	editCmd.Flags().BoolVar(&c.Edit.Watch, "watch", c.Edit.Watch, "Apply on save")

	return editCmd
}

func (c *Config) runEditCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if err := c.runEditor([]string{c.WorkingTreeAbsPath.String()}); err != nil {
			return err
		}
		if c.Edit.Apply {
			if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, noArgs, applyArgsOptions{
				cmd:          cmd,
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

	sourceState, err := c.newSourceState(cmd.Context(), cmd)
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
		preEditPlaintext []byte
	}
	var transparentlyDecryptedFiles []transparentlyDecryptedFile
TARGET_REL_PATH:
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		sourceRelPath := sourceStateEntry.SourceRelPath()
		switch sourceStateFile, ok := sourceStateEntry.(*chezmoi.SourceStateFile); {
		case ok && sourceStateFile.Attr().Encrypted:
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
				preEditPlaintext: contents,
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
			if sourceStateFile.Attr().Template {
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
			// not the case for most users. The user can set the tempDir
			// configuration variable if needed.
			if err := os.MkdirAll(hardlinkAbsPath.Dir().String(), 0o700); err != nil {
				return err
			}
			if err := c.baseSystem.Link(c.SourceDirAbsPath.Join(sourceRelPath.RelPath()), hardlinkAbsPath); err == nil {
				editorArgs = append(editorArgs, hardlinkAbsPath.String())
				continue TARGET_REL_PATH
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
			postEditPlaintext, err := c.baseSystem.ReadFile(transparentlyDecryptedFile.decryptedAbsPath)
			if err != nil {
				return err
			}
			if bytes.Equal(postEditPlaintext, transparentlyDecryptedFile.preEditPlaintext) {
				return nil
			}
			contents, err := c.encryption.EncryptFile(transparentlyDecryptedFile.decryptedAbsPath)
			if err != nil {
				return err
			}
			if err := c.baseSystem.WriteFile(transparentlyDecryptedFile.sourceAbsPath, contents, 0o666&^c.Umask); err != nil {
				return err
			}
		}

		if c.Edit.Apply || c.Edit.Watch {
			// Reset the cached source state to ensure that we re-read any
			// changed files.
			//
			// FIXME Be more precise in what we invalidate. Only the changed
			// files need to be re-read, not the entire source state.
			c.resetSourceState()

			if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
				cmd:          cmd,
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
					c.logger.Debug("watcher.Events", slog.String("Name", event.Name), chezmoilog.Stringer("Op", event.Op))
					err := postEditFunc()
					chezmoilog.InfoOrError(c.logger, "postEditFunc", err)
				case _, ok := <-watcher.Errors:
					if !ok {
						return
					}
					chezmoilog.InfoOrError(c.logger, "watcher.Errors", err)
				}
			}
		}()
	}

	if err := c.runEditor(editorArgs); err != nil {
		return err
	}

	return postEditFunc()
}
