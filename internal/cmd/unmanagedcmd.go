package cmd

import (
	"io/fs"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v3"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func (c *Config) newUnmanagedCmd() *cobra.Command {
	unmanagedCmd := &cobra.Command{
		Use:     "unmanaged",
		Short:   "List the unmanaged files in the destination directory",
		Long:    mustLongHelp("unmanaged"),
		Example: example("unmanaged"),
		Args:    cobra.NoArgs,
		RunE:    c.makeRunEWithSourceState(c.runUnmanagedCmd),
	}

	return unmanagedCmd
}

func (c *Config) runUnmanagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	sb := strings.Builder{}
	if err := chezmoi.WalkDir(c.destSystem, c.DestDirAbsPath, func(destAbsPath chezmoi.AbsPath, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if destAbsPath == c.DestDirAbsPath {
			return nil
		}
		targeRelPath := destAbsPath.MustTrimDirPrefix(c.DestDirAbsPath)
		_, managed := sourceState.Entry(targeRelPath)
		ignored := sourceState.Ignored(targeRelPath)
		if !managed && !ignored {
			sb.WriteString(string(targeRelPath) + "\n")
		}
		if info.IsDir() && (!managed || ignored) {
			return vfs.SkipDir
		}
		return nil
	}); err != nil {
		return err
	}
	return c.writeOutputString(sb.String())
}
