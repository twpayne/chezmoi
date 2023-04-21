package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type addCmdConfig struct {
	TemplateSymlinks bool `json:"templateSymlinks" mapstructure:"templateSymlinks" yaml:"templateSymlinks"`
	create           bool
	encrypt          bool
	exact            bool
	filter           *chezmoi.EntryTypeFilter
	follow           bool
	prompt           bool
	quiet            bool
	recursive        bool
	template         bool
}

func (c *Config) newAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:     "add targets...",
		Aliases: []string{"manage"},
		Short:   "Add an existing file, directory, or symlink to the source state",
		Long:    mustLongHelp("add"),
		Example: example("add"),
		Args:    cobra.MinimumNArgs(1),
		RunE:    c.makeRunEWithSourceState(c.runAddCmd),
		Annotations: newAnnotations(
			createSourceDirectoryIfNeeded,
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
			requiresWorkingTree,
		),
	}

	flags := addCmd.Flags()
	flags.BoolVar(&c.Add.create, "create", c.Add.create, "Add files that should exist, irrespective of their contents")
	flags.BoolVar(&c.Add.encrypt, "encrypt", c.Add.encrypt, "Encrypt files")
	flags.BoolVar(&c.Add.exact, "exact", c.Add.exact, "Add directories exactly")
	flags.VarP(c.Add.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.BoolVarP(&c.Add.follow, "follow", "f", c.Add.follow, "Add symlink targets instead of symlinks")
	flags.VarP(c.Add.filter.Include, "include", "i", "Include entry types")
	flags.BoolVarP(&c.Add.prompt, "prompt", "p", c.Add.prompt, "Prompt before adding each entry")
	flags.BoolVarP(&c.Add.quiet, "quiet", "q", c.Add.quiet, "Suppress warnings")
	flags.BoolVarP(&c.Add.recursive, "recursive", "r", c.Add.recursive, "Recurse into subdirectories")
	flags.BoolVarP(&c.Add.template, "template", "T", c.Add.template, "Add files as templates")
	flags.BoolVar(&c.Add.TemplateSymlinks, "template-symlinks", c.Add.TemplateSymlinks, "Add symlinks with target in source or home dirs as templates") //nolint:lll

	registerExcludeIncludeFlagCompletionFuncs(addCmd)

	return addCmd
}

func (c *Config) defaultOnIgnoreFunc(targetRelPath chezmoi.RelPath) {
	if !c.Add.quiet {
		c.errorf("warning: ignoring %s", targetRelPath)
	}
}

func (c *Config) defaultPreAddFunc(targetRelPath chezmoi.RelPath) error {
	if !c.Add.prompt {
		return nil
	}

	prompt := fmt.Sprintf("add %s", c.SourceDirAbsPath.Join(targetRelPath))
	for {
		switch choice, err := c.promptChoice(prompt, choicesYesNoAllQuit); {
		case err != nil:
			return err
		case choice == "all":
			c.Add.prompt = false
			return nil
		case choice == "no":
			return chezmoi.Skip
		case choice == "quit":
			return chezmoi.ExitCodeError(0)
		case choice == "yes":
			return nil
		default:
			panic(fmt.Sprintf("%s: unexpected choice", choice))
		}
	}
}

// defaultReplaceFunc prompts the user for confirmation if the adding the entry
// would remove any of the encrypted, private, or template attributes.
func (c *Config) defaultReplaceFunc(
	targetRelPath chezmoi.RelPath, newSourceStateEntry, oldSourceStateEntry chezmoi.SourceStateEntry,
) error {
	if c.force {
		return nil
	}

	newFile, newIsFile := newSourceStateEntry.(*chezmoi.SourceStateFile)
	oldFile, oldIsFile := oldSourceStateEntry.(*chezmoi.SourceStateFile)
	if !newIsFile || !oldIsFile {
		return nil
	}

	var removedAttributes []string
	if !newFile.Attr.Encrypted && oldFile.Attr.Encrypted {
		removedAttributes = append(removedAttributes, "encrypted")
	}
	if !newFile.Attr.Private && oldFile.Attr.Private {
		removedAttributes = append(removedAttributes, "private")
	}
	if !newFile.Attr.Template && oldFile.Attr.Template {
		removedAttributes = append(removedAttributes, "template")
	}
	if len(removedAttributes) == 0 {
		return nil
	}
	removedAttributesStr := englishListWithNoun(removedAttributes, "attribute", "")
	prompt := fmt.Sprintf("adding %s would remove %s, continue", targetRelPath, removedAttributesStr)

	for {
		switch choice, err := c.promptChoice(prompt, choicesYesNoAllQuit); {
		case err != nil:
			return err
		case choice == "all":
			c.force = true
			return nil
		case choice == "no":
			return chezmoi.Skip
		case choice == "quit":
			return chezmoi.ExitCodeError(0)
		case choice == "yes":
			return nil
		default:
			panic(fmt.Sprintf("%s: unexpected choice", choice))
		}
	}
}

func (c *Config) runAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	destAbsPathInfos, err := c.destAbsPathInfos(sourceState, args, destAbsPathInfosOptions{
		follow:    c.Mode == chezmoi.ModeSymlink || c.Add.follow,
		recursive: c.Add.recursive,
	})
	if err != nil {
		return err
	}

	persistentStateFileAbsPath, err := c.persistentStateFile()
	if err != nil {
		return err
	}

	return sourceState.Add(c.sourceSystem, c.persistentState, c.destSystem, destAbsPathInfos, &chezmoi.AddOptions{
		Create:          c.Add.create,
		Encrypt:         c.Add.encrypt,
		EncryptedSuffix: c.encryption.EncryptedSuffix(),
		Exact:           c.Add.exact,
		Filter:          c.Add.filter,
		OnIgnoreFunc:    c.defaultOnIgnoreFunc,
		PreAddFunc:      c.defaultPreAddFunc,
		ProtectedAbsPaths: []chezmoi.AbsPath{
			c.CacheDirAbsPath,
			c.WorkingTreeAbsPath,
			c.configFileAbsPath,
			persistentStateFileAbsPath,
			c.sourceDirAbsPath,
		},
		ReplaceFunc:      c.defaultReplaceFunc,
		Template:         c.Add.template,
		TemplateSymlinks: c.Add.TemplateSymlinks,
	})
}
