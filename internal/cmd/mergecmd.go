package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type mergeCmdConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

func (c *Config) newMergeCmd() *cobra.Command {
	mergeCmd := &cobra.Command{
		Use:     "merge target...",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Perform a three-way merge between the destination state, the source state, and the target state",
		Long:    mustLongHelp("merge"),
		Example: example("merge"),
		RunE:    c.makeRunEWithSourceState(c.runMergeCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			requiresSourceDirectory: "true",
		},
	}

	return mergeCmd
}

func (c *Config) runMergeCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: false,
		recursive:           true,
	})
	if err != nil {
		return err
	}

	// Create a temporary directory to store the target state and ensure that it
	// is removed afterwards. We cannot use fs as it lacks TempDir
	// functionality.
	tempDirAbsPath, err := c.tempDir("chezmoi-merge")
	if err != nil {
		return err
	}

	var plaintextTempDirAbsPath chezmoi.AbsPath
	defer func() {
		if !plaintextTempDirAbsPath.Empty() {
			_ = os.RemoveAll(plaintextTempDirAbsPath.String())
		}
	}()

	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)

		// If the source state entry is an encrypted file, then decrypt it to a
		// temporary directory and pass the plaintext to the merge command
		// instead.
		var (
			source           string
			plaintextAbsPath chezmoi.AbsPath
		)
		if sourceStateFile, ok := sourceStateEntry.(*chezmoi.SourceStateFile); ok {
			if sourceStateFile.Attr.Encrypted {
				plaintextTempDirAbsPath, err := c.tempDir("chezmoi-merge-plaintext")
				if err != nil {
					return err
				}
				plaintextAbsPath = plaintextTempDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath())
				plaintext, err := sourceStateFile.Contents()
				if err != nil {
					return err
				}
				if err := c.baseSystem.WriteFile(plaintextAbsPath, plaintext, 0o600); err != nil {
					return err
				}
				source = plaintextAbsPath.String()
			}
		}
		if source == "" {
			source = c.SourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath()).String()
		}

		// FIXME sourceStateEntry.TargetStateEntry eagerly evaluates the return
		// targetStateEntry's contents, which means that we cannot fallback to a
		// two-way merge if the source state's contents cannot be decrypted or
		// are an invalid template
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(c.destSystem, c.DestDirAbsPath.Join(targetRelPath))
		if err != nil {
			return fmt.Errorf("%s: %w", targetRelPath, err)
		}
		targetStateFile, ok := targetStateEntry.(*chezmoi.TargetStateFile)
		if !ok {
			// LATER consider handling symlinks?
			return fmt.Errorf("%s: not a file", targetRelPath)
		}
		contents, err := targetStateFile.Contents()
		if err != nil {
			return err
		}
		targetStateAbsPath := tempDirAbsPath.Join(chezmoi.RelPath(targetRelPath.Base()))
		if err := c.baseSystem.WriteFile(targetStateAbsPath, contents, 0o600); err != nil {
			return err
		}

		templateData := struct {
			Destination string
			Source      string
			Target      string
		}{
			Destination: c.DestDirAbsPath.Join(targetRelPath).String(),
			Source:      source,
			Target:      targetStateAbsPath.String(),
		}

		args := make([]string, 0, len(c.Merge.Args))
		// Work around a regression introduced in 2.1.4
		// (https://github.com/twpayne/chezmoi/pull/1324) in a user-friendly
		// way.
		//
		// Prior to #1324, the merge.args config option was prepended to the
		// default order of files to the merge command. Post #1324, the
		// merge.args config option replaced all arguments to the merge command.
		//
		// Work around this by looking for any templates in merge.args. An arg
		// is considered a template if, after execution as as template, it is
		// not equal to the original arg.
		anyTemplateArgs := false
		for i, arg := range c.Merge.Args {
			tmpl, err := template.New("merge.args[" + strconv.Itoa(i) + "]").Parse(arg)
			if err != nil {
				return err
			}

			var sb strings.Builder
			if err := tmpl.Execute(&sb, templateData); err != nil {
				return err
			}
			args = append(args, sb.String())

			// Detect template arguments.
			if arg != sb.String() {
				anyTemplateArgs = true
			}
		}

		if err := c.persistentState.Close(); err != nil {
			return err
		}

		// If there are no template arguments, then append the destination,
		// source, and target paths as prior to #1324.
		if !anyTemplateArgs {
			args = append(args, templateData.Destination, templateData.Source, templateData.Target)
		}

		if err := c.run(c.DestDirAbsPath, c.Merge.Command, args); err != nil {
			return fmt.Errorf("%s: %w", targetRelPath, err)
		}

		// If the source state entry was an encrypted file, then re-encrypt the
		// plaintext.
		if !plaintextAbsPath.Empty() {
			encryptedContents, err := c.encryption.EncryptFile(plaintextAbsPath)
			if err != nil {
				return err
			}
			if err := c.baseSystem.WriteFile(c.SourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath()), encryptedContents, 0o644); err != nil {
				return err
			}
		}
	}

	return nil
}
