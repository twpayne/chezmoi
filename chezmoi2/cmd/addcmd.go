package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type addCmdConfig struct {
	autoTemplate bool
	create       bool
	empty        bool
	encrypt      bool
	exact        bool
	follow       bool
	include      *chezmoi.IncludeSet
	recursive    bool
	template     bool
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
	flags.BoolVarP(&c.add.autoTemplate, "autotemplate", "a", c.add.autoTemplate, "auto generate the template when adding files as templates")
	flags.BoolVar(&c.add.create, "create", c.add.create, "add files that should exist, irrespective of their contents")
	flags.BoolVarP(&c.add.empty, "empty", "e", c.add.empty, "add empty files")
	flags.BoolVar(&c.add.encrypt, "encrypt", c.add.encrypt, "encrypt files")
	flags.BoolVarP(&c.add.exact, "exact", "x", c.add.exact, "add directories exactly")
	flags.BoolVarP(&c.add.follow, "follow", "f", c.add.follow, "add symlink targets instead of symlinks")
	flags.BoolVarP(&c.add.recursive, "recursive", "r", c.add.recursive, "recursive")
	flags.BoolVarP(&c.add.template, "template", "T", c.add.template, "add files as templates")

	return addCmd
}

func (c *Config) runAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	destAbsPathInfos, err := c.destAbsPathInfos(sourceState, args, c.add.recursive, c.add.follow)
	if err != nil {
		return err
	}

	return sourceState.Add(c.sourceSystem, c.persistentState, c.destSystem, destAbsPathInfos, &chezmoi.AddOptions{
		AutoTemplate:    c.add.autoTemplate,
		Create:          c.add.create,
		Empty:           c.add.empty,
		Encrypt:         c.add.encrypt,
		EncryptedSuffix: c.encryption.EncryptedSuffix(),
		Exact:           c.add.exact,
		Include:         c.add.include,
		Template:        c.add.template,
	})
}
