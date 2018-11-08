package cmd

import (
	"os"

	"github.com/absfs/afero"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var initCommand = &cobra.Command{
	Use:   "init",
	Args:  cobra.NoArgs,
	Short: "Initialize chezmoi",
	RunE:  makeRunE(runInitCommandE),
}

func init() {
	rootCommand.AddCommand(initCommand)
}

func runInitCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	actuator := config.getDefaultActuator(fs)
	fi, err := fs.Stat(config.SourceDir)
	switch {
	case err == nil && fi.Mode().IsDir():
		if fi.Mode()&os.ModePerm != 0700 {
			if err := actuator.Chmod(config.SourceDir, 0700); err != nil {
				return err
			}
		}
	case os.IsNotExist(err):
		if err := actuator.Mkdir(config.SourceDir, 0700); err != nil {
			return err
		}
	case err == nil:
		return errors.Errorf("%s: is not a directory", config.SourceDir)
	default:
		return err
	}
	return nil
}
