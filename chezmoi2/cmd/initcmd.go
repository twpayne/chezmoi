package cmd

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type initCmdConfig struct {
	apply         bool
	depth         int
	oneShot       bool
	purge         bool
	purgeBinary   bool
	skipEncrypted bool
}

var dotfilesRepoGuesses = []struct {
	rx     *regexp.Regexp
	format string
}{
	{
		rx:     regexp.MustCompile(`\A[-0-9A-Za-z]+\z`),
		format: "https://github.com/%s/dotfiles.git",
	},
	{
		rx:     regexp.MustCompile(`\A[-0-9A-Za-z]+/[-0-9A-Za-z]+\.git\z`),
		format: "https://github.com/%s",
	},
	{
		rx:     regexp.MustCompile(`\A[-0-9A-Za-z]+/[-0-9A-Za-z]+\z`),
		format: "https://github.com/%s.git",
	},
	{
		rx:     regexp.MustCompile(`\A[-.0-9A-Za-z]+/[-0-9A-Za-z]+\z`),
		format: "https://%s/dotfiles.git",
	},
	{
		rx:     regexp.MustCompile(`\A[-.0-9A-Za-z]+/[-0-9A-Za-z]+/[-0-9A-Za-z]+\z`),
		format: "https://%s.git",
	},
	{
		rx:     regexp.MustCompile(`\A[-.0-9A-Za-z]+/[-0-9A-Za-z]+/[-0-9A-Za-z]+\.git\z`),
		format: "https://%s",
	},
	{
		rx:     regexp.MustCompile(`\Asr\.ht/~[-0-9A-Za-z]+\z`),
		format: "https://git.%s/dotfiles",
	},
	{
		rx:     regexp.MustCompile(`\Asr\.ht/~[-0-9A-Za-z]+/[-0-9A-Za-z]+\z`),
		format: "https://git.%s",
	},
}

func (c *Config) newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Args:    cobra.MaximumNArgs(1),
		Use:     "init [repo]",
		Short:   "Setup the source directory and update the destination directory to match the target state",
		Long:    mustLongHelp("init"),
		Example: example("init"),
		RunE:    c.runInitCmd,
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			persistentStateMode:          persistentStateModeReadWrite,
			requiresSourceDirectory:      "true",
			runsCommands:                 "true",
		},
	}

	flags := initCmd.Flags()
	flags.BoolVarP(&c.init.apply, "apply", "a", c.init.apply, "update destination directory")
	flags.IntVarP(&c.init.depth, "depth", "d", c.init.depth, "create a shallow clone")
	flags.BoolVar(&c.init.oneShot, "one-shot", c.init.oneShot, "one shot")
	flags.BoolVarP(&c.init.purge, "purge", "p", c.init.purge, "purge config and source directories")
	flags.BoolVarP(&c.init.purgeBinary, "purge-binary", "P", c.init.purgeBinary, "purge chezmoi binary")
	flags.BoolVar(&c.init.skipEncrypted, "skip-encrypted", c.init.skipEncrypted, "skip encrypted files")

	return initCmd
}

func (c *Config) runInitCmd(cmd *cobra.Command, args []string) error {
	if c.init.oneShot {
		c.force = true
		c.init.apply = true
		c.init.depth = 1
		c.init.purge = true
	}

	if len(args) == 0 {
		switch useBuiltinGit, err := c.useBuiltinGit(); {
		case err != nil:
			return err
		case useBuiltinGit:
			rawSourceDir, err := c.baseSystem.RawPath(c.sourceDirAbsPath)
			if err != nil {
				return err
			}
			isBare := false
			_, err = git.PlainInit(string(rawSourceDir), isBare)
			return err
		default:
			return c.run(c.sourceDirAbsPath, c.Git.Command, []string{"init"})
		}
	}

	// Clone repo into source directory if it does not already exist.
	switch _, err := c.baseSystem.Stat(c.sourceDirAbsPath.Join(chezmoi.RelPath(".git"))); {
	case os.IsNotExist(err):
		rawSourceDir, err := c.baseSystem.RawPath(c.sourceDirAbsPath)
		if err != nil {
			return err
		}

		dotfilesRepoURL := guessDotfilesRepoURL(args[0])
		switch useBuiltinGit, err := c.useBuiltinGit(); {
		case err != nil:
			return err
		case useBuiltinGit:
			isBare := false
			if _, err := git.PlainClone(string(rawSourceDir), isBare, &git.CloneOptions{
				URL:               dotfilesRepoURL,
				Depth:             c.init.depth,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			}); err != nil {
				return err
			}
		default:
			args := []string{
				"clone",
				"--recurse-submodules",
			}
			if c.init.depth != 0 {
				args = append(args,
					"--depth", strconv.Itoa(c.init.depth),
				)
			}
			args = append(args,
				dotfilesRepoURL,
				string(rawSourceDir),
			)
			if err := c.run("", c.Git.Command, args); err != nil {
				return err
			}
		}
	case err != nil:
		return err
	}

	// Find config template, execute it, and create config file.
	filename, ext, data, err := c.findConfigTemplate()
	if err != nil {
		return err
	}
	var configFileContents []byte
	if filename != "" {
		configFileContents, err = c.createConfigFile(filename, data)
		if err != nil {
			return err
		}
	}

	// Reload config if it was created.
	if filename != "" {
		viper.SetConfigType(ext)
		if err := viper.ReadConfig(bytes.NewBuffer(configFileContents)); err != nil {
			return err
		}
		if err := viper.Unmarshal(c); err != nil {
			return err
		}
	}

	// Apply.
	if c.init.apply {
		if err := c.applyArgs(c.destSystem, c.destDirAbsPath, noArgs, applyArgsOptions{
			include:       chezmoi.NewIncludeSet(chezmoi.IncludeAll),
			recursive:     false,
			umask:         c.Umask.FileMode(),
			preApplyFunc:  c.defaultPreApplyFunc,
			skipEncrypted: c.init.skipEncrypted,
		}); err != nil {
			return err
		}
	}

	// Purge.
	if c.init.purge {
		if err := c.doPurge(&purgeOptions{
			binary: runtime.GOOS != "windows" && c.init.purgeBinary,
		}); err != nil {
			return err
		}
	}

	return nil
}

// createConfigFile creates a config file using a template and returns its
// contents.
func (c *Config) createConfigFile(filename chezmoi.RelPath, data []byte) ([]byte, error) {
	funcMap := make(template.FuncMap)
	for key, value := range c.templateFuncs {
		funcMap[key] = value
	}
	for name, f := range map[string]interface{}{
		"promptBool":   c.promptBool,
		"promptInt":    c.promptInt,
		"promptString": c.promptString,
	} {
		funcMap[name] = f
	}

	t, err := template.New(string(filename)).Funcs(funcMap).Parse(string(data))
	if err != nil {
		return nil, err
	}

	sb := strings.Builder{}
	if err = t.Execute(&sb, c.defaultTemplateData()); err != nil {
		return nil, err
	}
	contents := []byte(sb.String())

	configDir := chezmoi.AbsPath(c.bds.ConfigHome).Join("chezmoi")
	if err := chezmoi.MkdirAll(c.baseSystem, configDir, 0o777); err != nil {
		return nil, err
	}

	configPath := configDir.Join(filename)
	if err := c.baseSystem.WriteFile(configPath, contents, 0o600); err != nil {
		return nil, err
	}

	return contents, nil
}

func (c *Config) findConfigTemplate() (chezmoi.RelPath, string, []byte, error) {
	for _, ext := range viper.SupportedExts {
		filename := chezmoi.RelPath(chezmoi.Prefix + "." + ext + chezmoi.TemplateSuffix)
		contents, err := c.baseSystem.ReadFile(c.sourceDirAbsPath.Join(filename))
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return "", "", nil, err
		}
		return chezmoi.RelPath("chezmoi." + ext), ext, contents, nil
	}
	return "", "", nil, nil
}

func (c *Config) promptBool(field string) bool {
	value, err := parseBool(c.promptString(field))
	if err != nil {
		returnTemplateError(err)
		return false
	}
	return value
}

func (c *Config) promptInt(field string) int64 {
	value, err := strconv.ParseInt(c.promptString(field), 10, 64)
	if err != nil {
		returnTemplateError(err)
		return 0
	}
	return value
}

func (c *Config) promptString(field string) string {
	value, err := c.readLine(fmt.Sprintf("%s? ", field))
	if err != nil {
		returnTemplateError(err)
		return ""
	}
	return strings.TrimSpace(value)
}

// guessDotfilesRepoURL guesses the user's dotfile repo from arg.
func guessDotfilesRepoURL(arg string) string {
	for _, dotfileRepoGuess := range dotfilesRepoGuesses {
		if dotfileRepoGuess.rx.MatchString(arg) {
			return fmt.Sprintf(dotfileRepoGuess.format, arg)
		}
	}
	return arg
}
