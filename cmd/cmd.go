package cmd

import (
	"log"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

func makeRun(runCommand func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := runCommand(cmd, args); err != nil {
			log.Fatal(err)
		}
	}
}

func getUmask() os.FileMode {
	// FIXME should we call runtime.LockOSThread or similar?
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	return os.FileMode(umask)
}
