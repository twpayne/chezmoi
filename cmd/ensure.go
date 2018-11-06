package cmd

import (
	"os"
	"syscall"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var ensureCommand = &cobra.Command{
	Use:   "ensure",
	Short: "Ensure that the actual state matches the target state",
	Run:   makeRun(runEnsureCommand),
}

var (
	ensureVerbose = false
	ensureDryRun  = false
)

func init() {
	persistentFlags := ensureCommand.PersistentFlags()
	persistentFlags.BoolVar(&ensureVerbose, "verbose", ensureVerbose, "verbose")
	persistentFlags.BoolVar(&ensureDryRun, "dry-run", ensureDryRun, "Dry run")

	rootCommand.AddCommand(ensureCommand)
}

func runEnsureCommand(command *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	targetState, err := getTargetState(fs)
	if err != nil {
		return err
	}
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	var actuator chezmoi.Actuator
	if ensureDryRun {
		actuator = chezmoi.NewNullActuator()
	} else {
		actuator = chezmoi.NewFsActuator(fs)
	}
	if ensureVerbose {
		actuator = chezmoi.NewLoggingActuator(actuator)
	}
	return targetState.Ensure(fs, targetDir, os.FileMode(umask), actuator)
}
