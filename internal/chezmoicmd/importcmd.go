package chezmoicmd

// LATER add zip import

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type importCmdConfig struct {
	exclude           *chezmoi.EntryTypeSet
	destination       chezmoi.AbsPath
	exact             bool
	include           *chezmoi.EntryTypeSet
	removeDestination bool
	stripComponents   int
}

func (c *Config) newImportCmd() *cobra.Command {
	importCmd := &cobra.Command{
		Use:     "import archive",
		Short:   "Import a tar archive into the source state",
		Long:    mustLongHelp("import"),
		Example: example("import"),
		Args:    cobra.MaximumNArgs(1),
		RunE:    c.makeRunEWithSourceState(c.runImportCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			persistentStateMode:     persistentStateModeReadWrite,
			requiresSourceDirectory: "true",
		},
	}

	flags := importCmd.Flags()
	flags.VarP(&c._import.destination, "destination", "d", "destination prefix")
	flags.BoolVar(&c._import.exact, "exact", c._import.exact, "import directories exactly")
	flags.VarP(c._import.exclude, "exclude", "x", "exclude entry types")
	flags.VarP(c._import.include, "include", "i", "include entry types")
	flags.BoolVarP(&c._import.removeDestination, "remove-destination", "r", c._import.removeDestination, "remove destination before import")
	flags.IntVar(&c._import.stripComponents, "strip-components", c._import.stripComponents, "strip components")

	return importCmd
}

func (c *Config) runImportCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var r io.Reader
	if len(args) == 0 {
		r = c.stdin
	} else {
		absPath, err := chezmoi.NewAbsPathFromExtPath(args[0], c.homeDirAbsPath)
		if err != nil {
			return err
		}
		data, err := c.baseSystem.ReadFile(absPath)
		if err != nil {
			return err
		}
		r = bytes.NewReader(data)
		switch base := strings.ToLower(absPath.Base()); {
		case strings.HasSuffix(base, ".tar.gz") || strings.HasSuffix(base, ".tgz"):
			r, err = gzip.NewReader(r)
			if err != nil {
				return err
			}
		case strings.HasSuffix(base, ".tar.bz2"):
			r = bzip2.NewReader(r)
		case strings.HasSuffix(base, ".tar"):
		default:
			return fmt.Errorf("unknown format: %s", base)
		}
	}
	tarReaderSystem, err := chezmoi.NewTARReaderSystem(tar.NewReader(r), chezmoi.TARReaderSystemOptions{
		RootAbsPath:     c._import.destination,
		StripComponents: c._import.stripComponents,
	})
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
	return sourceState.Add(c.sourceSystem, c.persistentState, tarReaderSystem, tarReaderSystem.FileInfos(), &chezmoi.AddOptions{
		Exact:     c._import.exact,
		Include:   c._import.include.Sub(c._import.exclude),
		RemoveDir: removeDir,
	})
}
