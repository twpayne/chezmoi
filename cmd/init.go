package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"
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
command can automatically invoke the "apply" command to update the destination
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
	}

	if err := c.createConfigFile(fs); err != nil {
		return err
	}

	if c.init.apply {
		if err := c.applyArgs(fs, nil, mutator); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) createConfigFile(fs vfs.FS) error {
	templatePath, extension, err := c.findConfigTemplate(fs)
	if err != nil {
		return fmt.Errorf("failed to lookup config template: %v", err)
	}

	if templatePath == "" {
		// no config template file exists
		return nil
	}

	fmt.Fprintf(c.Stdout(), "Creating new configuration file from template %q\n", templatePath)
	configDir := c.bds.ConfigHome
	if _, err := fs.Stat(configDir); os.IsNotExist(err) {
		err = fs.Mkdir(configDir, 0775)
		if err != nil {
			return fmt.Errorf("failed to create config directory %q: %v", configDir, err)
		}
	}

	configDir = filepath.Join(configDir, "chezmoi")
	if _, err := fs.Stat(configDir); os.IsNotExist(err) {
		err = fs.Mkdir(configDir, 0775)
		if err != nil {
			return fmt.Errorf("failed to create chezmoi config directory %q: %v", configDir, err)
		}
	}

	data, err := fs.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read config template: %v", err)
	}

	t := template.New(templatePath)
	funcs := c.templateFuncs
	if funcs == nil {
		funcs = template.FuncMap{}
	}

	funcs["promptString"] = c.promptString
	t.Funcs(funcs)

	t, err = t.Parse(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse config template: %v", err)
	}

	contents := new(bytes.Buffer)
	err = t.Execute(contents, nil)
	if err != nil {
		return fmt.Errorf("failed to execute config template: %v", err)
	}

	configFile := filepath.Join(configDir, "chezmoi."+extension)

	if c.DryRun {
		fmt.Fprintf(c.Stdout(), "Would have written the following configuration file to %q\n", configFile)
		fmt.Fprint(c.Stdout(), contents.String())
		return nil
	} else {
		fmt.Fprintf(c.Stdout(), "Writing configuration file to %q\n", configFile)
		err = fs.WriteFile(configFile, contents.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}
	}

	// The last step is to reload the configuration so it can be used directly
	// if we want to apply the dotfiles. In unit tests the fs is typically a
	// *vfst.TestFS that actually operates in a temporary directory. Thus we
	// have to find the full absolute path to the configuration file now so
	// viper can load the correct configuration.
	if tfs, ok := fs.(*vfst.TestFS); ok {
		configFile = filepath.Join(tfs.TempDir(), configFile)
	}

	err = loadConfigFile(configFile, c)
	if err != nil {
		return fmt.Errorf("failed to reload config file: %v", err)
	}

	return nil
}

func (c *Config) findConfigTemplate(fs vfs.FS) (path, extension string, err error) {
	files, err := fs.ReadDir(c.SourceDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to open directory %q: %v", c.SourceDir, err)
	}

	extensions := viper.SupportedExts
	for i, ext := range extensions {
		extensions[i] = regexp.QuoteMeta(ext)
	}

	re := regexp.MustCompile(fmt.Sprintf(`^\.chezmoi\.(%s)\.tmpl$`,
		strings.Join(extensions, "|"),
	))

	for _, f := range files {
		name := f.Name()
		matches := re.FindStringSubmatch(name)
		if matches == nil {
			continue
		}

		ext := matches[1]
		return filepath.Join(c.SourceDir, name), ext, nil
	}

	return "", "", nil
}

func (c *Config) promptString(field string) string {
	reader := bufio.NewReader(c.Stdin())
	for {
		fmt.Fprintf(c.Stdout(), "Enter value for field %q: ", field)
		val, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Fprintf(c.Stdout(), "ERROR: failed to read input from stdin: EOF\n")
			os.Exit(1) // TODO?
		}
		if err != nil {
			fmt.Fprintf(c.Stdout(), "ERROR: %v: \n", err)
			continue
		}

		return val[:len(val)-1] // strip trailing newline
	}
}
