package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type initCmdConfig struct {
	apply       bool
	branch      string
	configPath  chezmoi.AbsPath
	data        bool
	depth       int
	exclude     *chezmoi.EntryTypeSet
	oneShot     bool
	purge       bool
	purgeBinary bool
	ssh         bool
}

var dotfilesRepoGuesses = []struct {
	rx            *regexp.Regexp
	httpGuessRepl string
	sshGuessRepl  string
}{
	{
		rx:            regexp.MustCompile(`\A([-0-9A-Za-z]+)\z`),
		httpGuessRepl: "https://github.com/$1/dotfiles.git",
		sshGuessRepl:  "git@github.com:$1/dotfiles.git",
	},
	{
		rx:            regexp.MustCompile(`\A([-0-9A-Za-z]+/[-0-9A-Za-z]+\.git)\z`),
		httpGuessRepl: "https://github.com/$1",
		sshGuessRepl:  "git@github.com:$1",
	},
	{
		rx:            regexp.MustCompile(`\A([-0-9A-Za-z]+/[-0-9A-Za-z]+)\z`),
		httpGuessRepl: "https://github.com/$1.git",
		sshGuessRepl:  "git@github.com:$1.git",
	},
	{
		rx:            regexp.MustCompile(`\A([-.0-9A-Za-z]+)/([-0-9A-Za-z]+)\z`),
		httpGuessRepl: "https://$1/$2/dotfiles.git",
		sshGuessRepl:  "git@$1:$2/dotfiles.git",
	},
	{
		rx:            regexp.MustCompile(`\A([-.0-9A-Za-z]+)/([-0-9A-Za-z]+/[-0-9A-Za-z]+)\z`),
		httpGuessRepl: "https://$1/$2.git",
		sshGuessRepl:  "git@$1:$2.git",
	},
	{
		rx:            regexp.MustCompile(`\A([-.0-9A-Za-z]+)/([-0-9A-Za-z]+/[-0-9A-Za-z]+\.git)\z`),
		httpGuessRepl: "https://$1/$2",
		sshGuessRepl:  "git@$1:$2",
	},
	{
		rx:            regexp.MustCompile(`\Asr\.ht/(~[-0-9A-Za-z]+)\z`),
		httpGuessRepl: "https://git.sr.ht/$1/dotfiles",
		sshGuessRepl:  "git@git.sr.ht:$1/dotfiles",
	},
	{
		rx:            regexp.MustCompile(`\Asr\.ht/(~[-0-9A-Za-z]+/[-0-9A-Za-z]+)\z`),
		httpGuessRepl: "https://git.sr.ht/$1",
		sshGuessRepl:  "git@git.sr.ht:$1",
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
	flags.VarP(&c.init.configPath, "config-path", "C", "Path to write generated config file")
	flags.BoolVar(&c.init.data, "data", c.init.data, "Include existing template data")
	flags.IntVarP(&c.init.depth, "depth", "d", c.init.depth, "Create a shallow clone")
	flags.VarP(c.init.exclude, "exclude", "x", "Exclude entry types")
	flags.BoolVar(&c.init.oneShot, "one-shot", c.init.oneShot, "Run in one-shot mode")
	flags.BoolVarP(&c.init.purge, "purge", "p", c.init.purge, "Purge config and source directories after running")
	flags.BoolVarP(&c.init.purgeBinary, "purge-binary", "P", c.init.purgeBinary, "Purge chezmoi binary after running")
	flags.StringVar(&c.init.branch, "branch", c.init.branch, "Set initial branch to checkout")
	flags.BoolVar(&c.init.ssh, "ssh", false, "Use ssh instead of https when guessing dotfile repo URL")

	return initCmd
}

func (c *Config) runInitCmd(cmd *cobra.Command, args []string) error {
	if c.init.oneShot {
		c.force = true
		c.init.apply = true
		c.init.depth = 1
		c.init.purge = true
		c.init.purgeBinary = true
	}

	// If the source repo does not exist then init or clone it.
	switch _, err := c.baseSystem.Stat(c.SourceDirAbsPath.Join(".git")); {
	case errors.Is(err, fs.ErrNotExist):
		rawSourceDir, err := c.baseSystem.RawPath(c.SourceDirAbsPath)
		if err != nil {
			return err
		}

		useBuiltinGit, err := c.UseBuiltinGit.Value()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if useBuiltinGit {
				isBare := false
				if _, err = git.PlainInit(string(rawSourceDir), isBare); err != nil {
					return err
				}
			} else if err := c.run(c.SourceDirAbsPath, c.Git.Command, []string{"init"}); err != nil {
				return err
			}
		} else {
			dotfilesRepoURL := guessDotfilesRepoURL(args[0], c.init.ssh)
			if useBuiltinGit {
				isBare := false
				var referenceName plumbing.ReferenceName
				if c.init.branch != "" {
					referenceName = plumbing.NewBranchReferenceName(c.init.branch)
				}
				cloneOptions := git.CloneOptions{
					URL:               dotfilesRepoURL,
					Depth:             c.init.depth,
					ReferenceName:     referenceName,
					RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				}

				if _, err := git.PlainClone(string(rawSourceDir), isBare, &cloneOptions); err != nil {
					if errors.Is(err, transport.ErrAuthenticationRequired) {
						reader := bufio.NewReader(os.Stdin)

						c.writeToStdout("Git authentication required - enter username and password...\n")
						c.writeToStdout("Username: ")
						username, err := reader.ReadString('\n')
						if err != nil {
							return err
						}
						c.writeToStdout("Password: ")
						passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
						c.writeToStdout("\n")
						if err != nil {
							return err
						}
						password := string(passwordBytes)
						username = strings.TrimSpace(username)
						password = strings.TrimSpace(password)
						cloneOptions.Auth = &http.BasicAuth{
							Username: username,
							Password: password,
						}
						if _, err := git.PlainClone(string(rawSourceDir), isBare, &cloneOptions); err != nil {
							return err
						}
					} else {
						return err
					}
				}
			} else {
				args := []string{
					"clone",
					"--recurse-submodules",
				}
				if c.init.branch != "" {
					args = append(args,
						"--branch", c.init.branch,
					)
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
		}
	case err != nil:
		return err
	}

	// Find config template, execute it, and create config file.
	configTemplateRelPath, ext, configTemplateContents, err := c.findConfigTemplate()
	if err != nil {
		return err
	}
	var configFileContents []byte
	if configTemplateRelPath == "" {
		if err := c.persistentState.Delete(chezmoi.ConfigStateBucket, configStateKey); err != nil {
			return err
		}
	} else {
		configFileContents, err = c.createConfigFile(configTemplateRelPath, configTemplateContents)
		if err != nil {
			return err
		}

		// Validate the config.
		v := viper.New()
		v.SetConfigType(ext)
		if err := v.ReadConfig(bytes.NewBuffer(configFileContents)); err != nil {
			return err
		}
		if err := v.Unmarshal(&Config{}, viperDecodeConfigOptions...); err != nil {
			return err
		}

		// Write the config.
		configPath := c.init.configPath
		if c.init.configPath == "" {
			configPath = chezmoi.AbsPath(c.bds.ConfigHome).Join("chezmoi").Join(configTemplateRelPath)
		}
		if err := chezmoi.MkdirAll(c.baseSystem, configPath.Dir(), 0o777); err != nil {
			return err
		}
		if err := c.baseSystem.WriteFile(configPath, configFileContents, 0o600); err != nil {
			return err
		}

		configStateValue, err := json.Marshal(configState{
			ConfigTemplateContentsSHA256: chezmoi.HexBytes(chezmoi.SHA256Sum(configTemplateContents)),
		})
		if err != nil {
			return err
		}
		if err := c.persistentState.Set(chezmoi.ConfigStateBucket, configStateKey, configStateValue); err != nil {
			return err
		}
	}

	// Reload config if it was created.
	if configTemplateRelPath != "" {
		v := viper.New()
		v.SetConfigType(ext)
		if err := v.ReadConfig(bytes.NewBuffer(configFileContents)); err != nil {
			return err
		}
		if err := v.Unmarshal(c, viperDecodeConfigOptions...); err != nil {
			return err
		}
	}

	// Apply.
	if c.init.apply {
		if err := c.applyArgs(c.destSystem, c.DestDirAbsPath, noArgs, applyArgsOptions{
			include:      chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll).Sub(c.init.exclude),
			recursive:    false,
			umask:        c.Umask,
			preApplyFunc: c.defaultPreApplyFunc,
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
	chezmoi.RecursiveMerge(funcMap, c.templateFuncs)
	chezmoi.RecursiveMerge(funcMap, map[string]interface{}{
		"promptBool":    c.promptBool,
		"promptInt":     c.promptInt,
		"promptString":  c.promptString,
		"stdinIsATTY":   c.stdinIsATTY,
		"writeToStdout": c.writeToStdout,
	})

	t, err := template.New(string(filename)).Funcs(funcMap).Parse(string(data))
	if err != nil {
		return nil, err
	}

	sb := strings.Builder{}
	templateData := c.defaultTemplateData()
	if c.init.data {
		chezmoi.RecursiveMerge(templateData, c.Data)
	}
	if err = t.Execute(&sb, templateData); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
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

func (c *Config) stdinIsATTY() bool {
	file, ok := c.stdin.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func (c *Config) writeToStdout(args ...string) string {
	for _, arg := range args {
		if _, err := c.stdout.Write([]byte(arg)); err != nil {
			panic(err)
		}
	}
	return ""
}

// guessDotfilesRepoURL guesses the user's dotfile repo from arg.
func guessDotfilesRepoURL(arg string, ssh bool) string {
	for _, dotfileRepoGuess := range dotfilesRepoGuesses {
		if !dotfileRepoGuess.rx.MatchString(arg) {
			continue
		}
		switch {
		case ssh && dotfileRepoGuess.sshGuessRepl != "":
			return dotfileRepoGuess.rx.ReplaceAllString(arg, dotfileRepoGuess.sshGuessRepl)
		case !ssh && dotfileRepoGuess.httpGuessRepl != "":
			return dotfileRepoGuess.rx.ReplaceAllString(arg, dotfileRepoGuess.httpGuessRepl)
		}
	}
	return arg
}
