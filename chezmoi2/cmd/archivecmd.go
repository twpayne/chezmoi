package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type archiveCmdConfig struct {
	format    string
	gzip      bool
	include   *chezmoi.IncludeSet
	recursive bool
}

func (c *Config) newArchiveCmd() *cobra.Command {
	archiveCmd := &cobra.Command{
		Use:     "archive [target]...",
		Short:   "Generate a tar archive of the target state",
		Long:    mustLongHelp("archive"),
		Example: example("archive"),
		RunE:    c.runArchiveCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeEmpty,
		},
	}

	flags := archiveCmd.Flags()
	flags.StringVar(&c.archive.format, "format", "tar", "format (tar or zip)")
	flags.BoolVarP(&c.archive.gzip, "gzip", "z", c.archive.gzip, "compress the output with gzip")
	flags.VarP(c.archive.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.archive.recursive, "recursive", "r", c.archive.recursive, "recursive")

	return archiveCmd
}

func (c *Config) runArchiveCmd(cmd *cobra.Command, args []string) error {
	output := strings.Builder{}
	var archiveSystem interface {
		chezmoi.System
		Close() error
	}
	switch c.archive.format {
	case "tar":
		archiveSystem = chezmoi.NewTARWriterSystem(&output, tarHeaderTemplate())
	case "zip":
		archiveSystem = chezmoi.NewZIPWriterSystem(&output, time.Now().UTC())
	default:
		return fmt.Errorf("%s: invalid format", c.archive.format)
	}
	if err := c.applyArgs(archiveSystem, "", args, applyArgsOptions{
		include:   c.archive.include,
		recursive: c.archive.recursive,
		umask:     os.ModePerm,
	}); err != nil {
		return err
	}
	if err := archiveSystem.Close(); err != nil {
		return err
	}

	if c.archive.format == "zip" || !c.archive.gzip {
		return c.writeOutputString(output.String())
	}

	gzippedArchive := strings.Builder{}
	w := gzip.NewWriter(&gzippedArchive)
	if _, err := w.Write([]byte(output.String())); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
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
