package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type importCmdConfig struct {
	filter            *chezmoi.EntryTypeFilter
	destination       chezmoi.AbsPath
	stripComponents   int
	exact             bool
	removeDestination bool
}

func (c *Config) newImportCmd() *cobra.Command {
	importCmd := &cobra.Command{
		Use:     "import archive",
		Short:   "Import an archive into the source state",
		Long:    mustLongHelp("import"),
		Example: example("import"),
		Args:    cobra.MaximumNArgs(1),
		RunE:    c.makeRunEWithSourceState(c.runImportCmd),
		Annotations: newAnnotations(
			createSourceDirectoryIfNeeded,
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}

	flags := importCmd.Flags()
	flags.VarP(&c._import.destination, "destination", "d", "Set destination prefix")
	flags.BoolVar(
		&c._import.exact,
		"exact",
		c._import.exact,
		"Set exact_ attribute on imported directories",
	)
	flags.VarP(c._import.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c._import.filter.Include, "include", "i", "Include entry types")
	flags.BoolVarP(
		&c._import.removeDestination,
		"remove-destination",
		"r",
		c._import.removeDestination,
		"Remove destination before import",
	)
	flags.IntVar(
		&c._import.stripComponents,
		"strip-components",
		c._import.stripComponents,
		"Strip leading path components",
	)

	registerExcludeIncludeFlagCompletionFuncs(importCmd)

	return importCmd
}

func (c *Config) runImportCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var (
		name string
		data []byte
	)
	if len(args) == 0 {
		name = ".tar"
		var err error
		data, err = io.ReadAll(c.stdin)
		if err != nil {
			return err
		}
	} else {
		absPath, err := chezmoi.NewAbsPathFromExtPath(args[0], c.homeDirAbsPath)
		if err != nil {
			return err
		}
		name = absPath.String()
		data, err = c.baseSystem.ReadFile(absPath)
		if err != nil {
			return err
		}
	}
	archiveReaderSystem, err := chezmoi.NewArchiveReaderSystem(
		name, data, chezmoi.ArchiveFormatUnknown, chezmoi.ArchiveReaderSystemOptions{
			RootAbsPath:     c._import.destination,
			StripComponents: c._import.stripComponents,
		},
	)
	if err != nil {
		return err
	}
	var removeDir chezmoi.RelPath
	if c._import.removeDestination {
		removeDir, err = c._import.destination.TrimDirPrefix(c.DestDirAbsPath)
		if err != nil {
			return err
		}
	}
	return sourceState.Add(
		c.sourceSystem,
		c.persistentState,
		archiveReaderSystem,
		archiveReaderSystem.FileInfos(),
		&chezmoi.AddOptions{
			Exact:     c._import.exact,
			Filter:    c._import.filter,
			RemoveDir: removeDir,
		},
	)
}
