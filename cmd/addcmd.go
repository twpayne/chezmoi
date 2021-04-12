package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type addCmdConfig struct {
	TemplateSymlinks bool `mapstructure:"templateSymlinks"`
	autoTemplate     bool
	create           bool
	empty            bool
	encrypt          bool
	exact            bool
	exclude          *chezmoi.EntryTypeSet
	follow           bool
	include          *chezmoi.EntryTypeSet
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
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			persistentStateMode:     persistentStateModeReadWrite,
			requiresSourceDirectory: "true",
		},
	}

	flags := addCmd.Flags()
	flags.BoolVarP(&c.Add.autoTemplate, "autotemplate", "a", c.Add.autoTemplate, "auto generate the template when adding files as templates")
	flags.BoolVar(&c.Add.create, "create", c.Add.create, "add files that should exist, irrespective of their contents")
	flags.BoolVarP(&c.Add.empty, "empty", "e", c.Add.empty, "add empty files")
	flags.BoolVar(&c.Add.encrypt, "encrypt", c.Add.encrypt, "encrypt files")
	flags.BoolVar(&c.Add.exact, "exact", c.Add.exact, "add directories exactly")
	flags.VarP(c.Add.exclude, "exclude", "x", "exclude entry types")
	flags.BoolVarP(&c.Add.follow, "follow", "f", c.Add.follow, "add symlink targets instead of symlinks")
	flags.BoolVarP(&c.Add.recursive, "recursive", "r", c.Add.recursive, "recursive")
	flags.BoolVarP(&c.Add.template, "template", "T", c.Add.template, "add files as templates")
	flags.BoolVar(&c.Add.TemplateSymlinks, "template-symlinks", c.Add.TemplateSymlinks, "add symlinks with target in source or home dirs as templates")

	return addCmd
}

func (c *Config) runAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	destAbsPathInfos, err := c.destAbsPathInfos(sourceState, args, c.Add.recursive, c.Add.follow)
	if err != nil {
		return err
	}

	return sourceState.Add(c.sourceSystem, c.persistentState, c.destSystem, destAbsPathInfos, &chezmoi.AddOptions{
		AutoTemplate:     c.Add.autoTemplate,
		Create:           c.Add.create,
		Empty:            c.Add.empty,
		Encrypt:          c.Add.encrypt,
		EncryptedSuffix:  c.encryption.EncryptedSuffix(),
		Exact:            c.Add.exact,
		Include:          c.Add.include.Sub(c.Add.exclude),
		Template:         c.Add.template,
		TemplateSymlinks: c.Add.TemplateSymlinks,
	})
}
