package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	"github.com/twpayne/go-vfs"
	"os"
	"path/filepath"
)

// initCmd represents the init command
var initCommand = &cobra.Command{
	Args:  cobra.ExactArgs(1),
	Use:   "init <repo URL>",
	Short: "Initial setup of the source directory then update the destination directory to match the target state",
	Long: `Initial setup of the source directory then update the destination directory to match the target state.

This command is supposed to run once when you want to setup your dotfiles on a
new host. It will clone the given repository into your source directory (see --source flag)
and make sure that all directory permissions are correct.

After your source directory was checked out and setup (e.g. git submodules) this
command will automatically invoke the "apply" command to update the destination
directory. You can use the --no-apply flag to prevent this from happening.
`,
	Example: `
  # Checkout from github using the public HTTPS API
  chezmoi init https://github.com/example/dotfiles.git

  # Checkout from github using your private key
  chezmoi init git@github.com:example/dotfiles.git
`,
	RunE: makeRunE(config.runInitCommand),
}

type initCommandConfig struct {
	noApply bool
}

func init() {
	rootCommand.AddCommand(initCommand)

	persistentFlags := initCommand.PersistentFlags()
	persistentFlags.BoolVar(&config.init.noApply, "no-apply", false, "do not update destination directory")
}

func (c *Config) runInitCommand(fs vfs.FS, args []string) error {
	mutator := c.getDefaultMutator(fs)
	err := c.initSourceParentDirs(fs, mutator)
	if err != nil {
		return err
	}

	err = c.checkoutSourceDir(fs, mutator, args[0])
	if err != nil {
		return err
	}

	if config.init.noApply {
		fmt.Println("Did not update target directory (--no-apply)")
		return nil
	}

	fmt.Printf("Updating target directory %q\n", c.DestDir)
	ts, err := c.getTargetState(fs)
	return ts.Apply(fs, mutator)
}

func (c *Config) initSourceParentDirs(fs vfs.FS, mutator chezmoi.Mutator) error {
	dirMode := os.FileMode(0700)
	parts := splitPath(c.SourceDir)
	parts = parts[:len(parts)-1] // all parent directories of the source directory

	var createdDir string
	for _, path := range parts {
		_, err := fs.Stat(path)
		switch {
		case os.IsNotExist(err):
			createdDir = path
			err = mutator.Mkdir(path, dirMode)
			if err != nil {
				return err
			}
		case err != nil:
			return err
		}
	}

	if createdDir != "" {
		fmt.Printf("Created source directory %q with file mode 0%o (%v) \n", createdDir, dirMode, dirMode)
	}

	return nil
}

func (c *Config) checkoutSourceDir(fs vfs.FS, mutator chezmoi.Mutator, repo string) error {
	fmt.Printf("Checking out dotfiles repository %q into %q\n", repo, c.SourceDir)

	vcsCmd := c.SourceVCSCommand
	if vcsCmd == "" {
		vcsCmd = "git"
	}

	var args []string
	switch vcsCmd {
	case "git", "hg":
		args = []string{"clone", repo, c.SourceDir}
	case "test":
		// While "test" isn't actually checking out anything, this command is
		// used to unit test this function and potentially for debugging.
		vcsCmd = "echo"
		args = []string{"-n", "clone", repo, c.SourceDir}
		err := mutator.Mkdir(c.SourceDir, 0777)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("VCS command %q is not yet support in chezmoi init", c.SourceVCSCommand)
	}

	err := c.execCmd(vcsCmd, args...)
	if err != nil {
		return err
	}

	err = mutator.Chmod(c.SourceDir, 0700)
	if err != nil {
		return err
	}

	if info, err := fs.Stat(filepath.Join(c.SourceDir, ".gitmodules")); err == nil && !info.IsDir() {
		return c.initSourceSubmodules()
	}

	return nil
}

func (c *Config) initSourceSubmodules() error {
	fmt.Printf("Initializing git submodules in %q\n", c.SourceDir)

	if c.Verbose {
		fmt.Printf("cd %s\n", c.SourceDir)
	}

	if !c.DryRun {
		if err := os.Chdir(c.SourceDir); err != nil {
			return err
		}
	}

	err := c.execCmd("git", "submodule", "init")
	if err != nil {
		return err
	}

	return c.execCmd("git", "submodule", "update")
}

func splitPath(p string) []string {
	dirs := []string{p}

	dir := filepath.Dir(p)
	for dir != string(filepath.Separator) {
		dirs = append(dirs, dir)
		dir = filepath.Dir(dir)
	}

	// reverse dirs slice (i.e. order by path length in ascending order)
	for i := 0; i < len(dirs)/2; i++ {
		j := len(dirs) - i - 1 // index from the end of dirs
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	return dirs
}
