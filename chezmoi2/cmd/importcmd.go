package cmd

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

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type importCmdConfig struct {
	destination       string
	exact             bool
	include           *chezmoi.IncludeSet
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
	flags.StringVarP(&c._import.destination, "destination", "d", c._import.destination, "destination prefix")
	flags.BoolVarP(&c._import.exact, "exact", "x", c._import.exact, "import directories exactly")
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
	rootAbsPath, err := chezmoi.NewAbsPathFromExtPath(c._import.destination, c.homeDirAbsPath)
	if err != nil {
		return err
	}
	tarReaderSystem, err := chezmoi.NewTARReaderSystem(tar.NewReader(r), chezmoi.TARReaderSystemOptions{
		RootAbsPath:     rootAbsPath,
		StripComponents: c._import.stripComponents,
	})
	if err != nil {
		return err
	}
	var removeDir chezmoi.RelPath
	if c._import.removeDestination {
		removeDir, err = rootAbsPath.TrimDirPrefix(c.destDirAbsPath)
		if err != nil {
			return err
		}
	}
	return sourceState.Add(c.sourceSystem, c.persistentState, tarReaderSystem, tarReaderSystem.FileInfos(), &chezmoi.AddOptions{
		Exact:     c._import.exact,
		Include:   c._import.include,
		RemoveDir: removeDir,
	})
}
