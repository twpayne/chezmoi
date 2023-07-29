package cmd

import (
	"archive/tar"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type archiveCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
	format    chezmoi.ArchiveFormat
	gzip      bool
	init      bool
	recursive bool
}

func (c *Config) newArchiveCmd() *cobra.Command {
	archiveCmd := &cobra.Command{
		Use:               "archive [target]...",
		Short:             "Generate a tar archive of the target state",
		Long:              mustLongHelp("archive"),
		Example:           example("archive"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runArchiveCmd,
		Annotations: newAnnotations(
			persistentStateModeEmpty,
			requiresSourceDirectory,
		),
	}

	flags := archiveCmd.Flags()
	flags.VarP(c.archive.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(&c.archive.format, "format", "f", "Set archive format")
	flags.BoolVarP(&c.archive.gzip, "gzip", "z", c.archive.gzip, "Compress output with gzip")
	flags.VarP(c.archive.filter.Exclude, "include", "i", "Include entry types")
	flags.BoolVar(&c.archive.init, "init", c.archive.init, "Recreate config file from template")
	flags.BoolVarP(
		&c.archive.recursive,
		"recursive",
		"r",
		c.archive.recursive,
		"Recurse into subdirectories",
	)

	registerExcludeIncludeFlagCompletionFuncs(archiveCmd)

	return archiveCmd
}

func (c *Config) runArchiveCmd(cmd *cobra.Command, args []string) error {
	format := c.archive.format
	if format == chezmoi.ArchiveFormatUnknown {
		format = chezmoi.GuessArchiveFormat(c.outputAbsPath.String(), nil)
		if format == chezmoi.ArchiveFormatUnknown {
			format = chezmoi.ArchiveFormatTar
		}
	}

	gzipOutput := c.archive.gzip
	if format == chezmoi.ArchiveFormatTarGz || format == chezmoi.ArchiveFormatTgz {
		gzipOutput = true
	}

	output := strings.Builder{}
	var archiveSystem interface {
		chezmoi.System
		Close() error
	}
	switch format {
	case chezmoi.ArchiveFormatTar, chezmoi.ArchiveFormatTarGz, chezmoi.ArchiveFormatTgz:
		archiveSystem = chezmoi.NewTarWriterSystem(&output, tarHeaderTemplate())
	case chezmoi.ArchiveFormatZip:
		archiveSystem = chezmoi.NewZIPWriterSystem(&output, time.Now().UTC())
	default:
		return chezmoi.UnknownArchiveFormatError(format)
	}
	if err := c.applyArgs(cmd.Context(), archiveSystem, chezmoi.EmptyAbsPath, args, applyArgsOptions{
		cmd:       cmd,
		filter:    c.archive.filter,
		init:      c.archive.init,
		recursive: c.archive.recursive,
	}); err != nil {
		return err
	}
	if err := archiveSystem.Close(); err != nil {
		return err
	}

	if format == chezmoi.ArchiveFormatZip || !gzipOutput {
		return c.writeOutputString(output.String())
	}

	gzippedArchive := strings.Builder{}
	gzipWriter := gzip.NewWriter(&gzippedArchive)
	if _, err := gzipWriter.Write([]byte(output.String())); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}
	return c.writeOutputString(gzippedArchive.String())
}

// tarHeaderTemplate returns a tar.Header template populated with the current
// user and time.
func tarHeaderTemplate() tar.Header {
	// Attempt to lookup the current user. Ignore errors because the default
	// zero values are reasonable.
	var (
		uid   int
		gid   int
		uname string
		gname string
	)
	if currentUser, err := user.Current(); err == nil {
		uid, _ = strconv.Atoi(currentUser.Uid)
		gid, _ = strconv.Atoi(currentUser.Gid)
		uname = currentUser.Username
		if group, err := user.LookupGroupId(currentUser.Gid); err == nil {
			gname = group.Name
		}
	}

	now := time.Now().UTC()
	return tar.Header{
		Uid:        uid,
		Gid:        gid,
		Uname:      uname,
		Gname:      gname,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
}
