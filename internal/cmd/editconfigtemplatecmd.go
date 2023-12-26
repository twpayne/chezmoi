package cmd

import (
	"errors"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func (c *Config) newEditConfigTemplateCmd() *cobra.Command {
	editConfigCmd := &cobra.Command{
		Use:     "edit-config-template",
		Short:   "Edit the configuration file template",
		Long:    mustLongHelp("edit-config-template"),
		Example: example("edit-config-template"),
		Args:    cobra.NoArgs,
		RunE:    c.makeRunEWithSourceState(c.runEditConfigTemplateCmd),
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			modifiesSourceDirectory,
			runsCommands,
		),
	}

	return editConfigCmd
}

func (c *Config) runEditConfigTemplateCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var configTemplateAbsPath chezmoi.AbsPath
	switch configTemplate, err := c.findConfigTemplate(); {
	case err != nil:
		return err
	case configTemplate != nil:
		configTemplateAbsPath = configTemplate.sourceAbsPath
	default:
		if err := chezmoi.MkdirAll(c.sourceSystem, c.sourceDirAbsPath, fs.ModePerm); err != nil &&
			!errors.Is(err, fs.ErrExist) {
			return err
		}
		configFileBase := "." + c.configFileAbsPath.Base() + ".tmpl"
		configTemplateAbsPath = c.sourceDirAbsPath.JoinString(configFileBase)
		switch data, err := c.baseSystem.ReadFile(c.configFileAbsPath); {
		case errors.Is(err, fs.ErrNotExist):
			// Do nothing.
		case err != nil:
			return err
		default:
			if err := c.sourceSystem.WriteFile(configTemplateAbsPath, data, 0o666); err != nil {
				return err
			}
		}
	}
	return c.runEditor([]string{configTemplateAbsPath.String()})
}
