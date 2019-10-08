package cmd

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
	"github.com/twpayne/chezmoi/internal/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var _importCmd = &cobra.Command{
	Use:     "import [filename]",
	Args:    cobra.MaximumNArgs(1),
	Short:   "Import a tar archive into the source state",
	Long:    mustGetLongHelp("import"),
	Example: getExample("import"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runImportCmd),
}

type importCmdConfig struct {
	removeDestination bool
	importTAROptions  chezmoi.ImportTAROptions
}

func init() {
	rootCmd.AddCommand(_importCmd)

	persistentFlags := _importCmd.PersistentFlags()
	persistentFlags.StringVarP(&config._import.importTAROptions.DestinationDir, "destination", "d", "", "destination prefix")
	persistentFlags.BoolVarP(&config._import.importTAROptions.Exact, "exact", "x", false, "import directories exactly")
	persistentFlags.IntVar(&config._import.importTAROptions.StripComponents, "strip-components", 0, "strip components")
	persistentFlags.BoolVarP(&config._import.removeDestination, "remove-destination", "r", false, "remove destination before import")
}

func (c *Config) runImportCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs, nil)
	if err != nil {
		return err
	}
	var r io.Reader
	if len(args) == 0 {
		r = c.Stdin()
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
		entry, err := ts.Get(fs, c._import.importTAROptions.DestinationDir)
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
