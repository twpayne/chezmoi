package cmd

// FIXME add zip support
// FIXME add --apply flag
// FIXME add --diff flag
// FIXME add --prompt flag

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var _importCommand = &cobra.Command{
	Use:   "import",
	Args:  cobra.MaximumNArgs(1),
	Short: "Import an archive",
	RunE:  makeRunE(config.runImportCommand),
}

type importCommandConfig struct {
	removeDestination bool
	importTAROptions  chezmoi.ImportTAROptions
}

func init() {
	rootCommand.AddCommand(_importCommand)

	persistentFlags := _importCommand.PersistentFlags()
	persistentFlags.StringVarP(&config._import.importTAROptions.DestinationDir, "destination", "d", "", "destination prefix")
	persistentFlags.BoolVarP(&config._import.importTAROptions.Exact, "exact", "x", false, "import directories exactly")
	persistentFlags.IntVar(&config._import.importTAROptions.StripComponents, "strip-components", 0, "strip components")
	persistentFlags.BoolVarP(&config._import.removeDestination, "remove-destination", "r", false, "remove destination before import")
}

func (c *Config) runImportCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	var r io.Reader
	if len(args) == 0 {
		r = os.Stdin
	} else {
		arg := args[0]
		f, err := fs.Open(arg)
		if err != nil {
			return err
		}
		defer f.Close()
		switch {
		case strings.HasSuffix(arg, ".tar.gz") || strings.HasSuffix(arg, ".tgz"):
			r, err = gzip.NewReader(f)
			if err != nil {
				return err
			}
		case strings.HasSuffix(arg, ".tar.bz2"):
			r = bzip2.NewReader(f)
		case strings.HasSuffix(arg, ".tar"):
			r = f
		default:
			return fmt.Errorf("%s: unknown format", arg)
		}
	}
	mutator := c.getDefaultMutator(fs)
	if c._import.removeDestination {
		entry, err := ts.Get(c._import.importTAROptions.DestinationDir)
		switch {
		case err == nil:
			if err := mutator.RemoveAll(filepath.Join(c.SourceDir, entry.SourceName())); err != nil {
				return err
			}
		case os.IsNotExist(err):
		default:
			return err
		}
	}
	return ts.ImportTAR(tar.NewReader(r), c._import.importTAROptions, mutator)
}
