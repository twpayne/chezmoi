package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var initCmd = &cobra.Command{
	Args:  cobra.MaximumNArgs(1),
	Use:   "init [repo]",
	Short: "Initial setup of the source directory then update the destination directory to match the target state",
	Long: `Initial setup of the source directory then update the destination directory to match the target state.

This command is supposed to run once when you want to setup your dotfiles on a
new host. It will clone the given repository into your source directory (see --source flag)
and make sure that all directory permissions are correct.

After your source directory was checked out and setup (e.g. git submodules) this
command will can automatically invoke the "apply" command to update the destination
directory if you supply the flag.
`,
	Example: `
  # Checkout from github using the public HTTPS API
  chezmoi init https://github.com/example/dotfiles.git

  # Checkout from github using your private key
  chezmoi init git@github.com:example/dotfiles.git
`,
	RunE: makeRunE(config.runInitCmd),
}

type initCmdConfig struct {
	apply bool
}

func init() {
	rootCmd.AddCommand(initCmd)

	persistentFlags := initCmd.PersistentFlags()
	persistentFlags.BoolVar(&config.init.apply, "apply", false, "update destination directory")
}

func (c *Config) runInitCmd(fs vfs.FS, args []string) error {
	vcsInfo, err := c.getVCSInfo()
	if err != nil {
		return err
	}

	mutator := c.getDefaultMutator(fs)

	if err := c.ensureSourceDirectory(fs, mutator); err != nil {
		return err
	}

	switch len(args) {
	case 0: // init
		var initArgs []string
		if c.SourceVCS.Init != nil {
			switch v := c.SourceVCS.Init.(type) {
			case string:
				initArgs = strings.Split(v, " ")
			case []string:
				initArgs = v
			default:
				return fmt.Errorf("sourceVCS.init: cannot parse value")
			}
		} else {
			initArgs = vcsInfo.initArgs
		}
		if err := c.run(c.SourceDir, c.SourceVCS.Command, initArgs...); err != nil {
			return err
		}
	case 1: // clone
		if vcsInfo.cloneArgsFunc == nil {
			return fmt.Errorf("%s: cloning not supported", c.SourceVCS.Command)
		}
		cloneArgs := vcsInfo.cloneArgsFunc(args[0], c.SourceDir)
		if err := c.run("", c.SourceVCS.Command, cloneArgs...); err != nil {
			return err
		}
		// FIXME this should be part of struct vcs
		switch filepath.Base(c.SourceVCS.Command) {
		case "git":
			if _, err := fs.Stat(filepath.Join(c.SourceDir, ".gitmodules")); err == nil {
				for _, args := range [][]string{
					[]string{"submodule", "init"},
					[]string{"submodule", "update"},
				} {
					if err := c.run(c.SourceDir, c.SourceVCS.Command, args...); err != nil {
						return err
					}
				}
			}
		}
		if c.init.apply {
			if err := c.applyArgs(fs, nil, mutator); err != nil {
				return err
			}
		}
	}

	return nil
}
