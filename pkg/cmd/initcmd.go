package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"runtime"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type initCmdConfig struct {
	apply           bool
	branch          string
	configPath      chezmoi.AbsPath
	data            bool
	depth           int
	exclude         *chezmoi.EntryTypeSet
	guessRepoURL    bool
	oneShot         bool
	forcePromptOnce bool
	promptBool      map[string]string
	promptInt       map[string]int
	promptString    map[string]string
	purge           bool
	purgeBinary     bool
	ssh             bool
}

var dotfilesRepoGuesses = []struct {
	rx                    *regexp.Regexp
	httpRepoGuessRepl     string
	httpUsernameGuessRepl string
	sshRepoGuessRepl      string
}{
	{
		rx:                    regexp.MustCompile(`\A([-0-9A-Za-z]+)\z`),
		httpRepoGuessRepl:     "https://github.com/$1/dotfiles.git",
		httpUsernameGuessRepl: "$1",
		sshRepoGuessRepl:      "git@github.com:$1/dotfiles.git",
	},
	{
		rx:                    regexp.MustCompile(`\A([-0-9A-Za-z]+)/([-0-9A-Za-z]+)(\.git)?\z`),
		httpRepoGuessRepl:     "https://github.com/$1/$2.git",
		httpUsernameGuessRepl: "$1",
		sshRepoGuessRepl:      "git@github.com:$1/$2.git",
	},
	{
		rx:                    regexp.MustCompile(`\A([-.0-9A-Za-z]+)/([-0-9A-Za-z]+)\z`),
		httpRepoGuessRepl:     "https://$1/$2/dotfiles.git",
		httpUsernameGuessRepl: "$2",
		sshRepoGuessRepl:      "git@$1:$2/dotfiles.git",
	},
	{
		rx:                    regexp.MustCompile(`\A([-0-9A-Za-z]+)/([-0-9A-Za-z]+)/([-.0-9A-Za-z]+)\z`),
		httpRepoGuessRepl:     "https://$1/$2/$3.git",
		httpUsernameGuessRepl: "$2",
		sshRepoGuessRepl:      "git@$1:$2/$3.git",
	},
	{
		rx:                    regexp.MustCompile(`\A([-.0-9A-Za-z]+)/([-0-9A-Za-z]+)/([-0-9A-Za-z]+)(\.git)?\z`),
		httpRepoGuessRepl:     "https://$1/$2/$3.git",
		httpUsernameGuessRepl: "$2",
		sshRepoGuessRepl:      "git@$1:$2/$3.git",
	},
	{
		rx:                    regexp.MustCompile(`\A(https?://)([-.0-9A-Za-z]+)/([-0-9A-Za-z]+)/([-0-9A-Za-z]+)(\.git)?\z`),
		httpRepoGuessRepl:     "$1$2/$3/$4.git",
		httpUsernameGuessRepl: "$3",
		sshRepoGuessRepl:      "git@$2:$3/$4.git",
	},
	{
		rx:                    regexp.MustCompile(`\Asr\.ht/~([-0-9A-Za-z]+)\z`),
		httpRepoGuessRepl:     "https://git.sr.ht/~$1/dotfiles",
		httpUsernameGuessRepl: "$1",
		sshRepoGuessRepl:      "git@git.sr.ht:~$1/dotfiles",
	},
	{
		rx:                    regexp.MustCompile(`\Asr\.ht/~([-0-9A-Za-z]+)/([-0-9A-Za-z]+)\z`),
		httpRepoGuessRepl:     "https://git.sr.ht/~$1/$2",
		httpUsernameGuessRepl: "$1",
		sshRepoGuessRepl:      "git@git.sr.ht:~$1/$2",
	},
}

// A loggableGitCloneOptions is a git.CloneOptions that implements
// github.com/rs/zerolog.LogObjectMarshaler.
type loggableGitCloneOptions git.CloneOptions

func (c *Config) newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Args:    cobra.MaximumNArgs(1),
		Use:     "init [repo]",
		Short:   "Setup the source directory and update the destination directory to match the target state",
		Long:    mustLongHelp("init"),
		Example: example("init"),
		RunE:    c.runInitCmd,
		Annotations: map[string]string{
			createSourceDirectoryIfNeeded: "true",
			modifiesDestinationDirectory:  "true",
			persistentStateMode:           persistentStateModeReadWrite,
			requiresWorkingTree:           "true",
			runsCommands:                  "true",
		},
	}

	flags := initCmd.Flags()
	flags.BoolVarP(&c.init.apply, "apply", "a", c.init.apply, "Update destination directory")
	flags.StringVar(&c.init.branch, "branch", c.init.branch, "Set initial branch to checkout")
	flags.VarP(&c.init.configPath, "config-path", "C", "Path to write generated config file")
	flags.BoolVar(&c.init.data, "data", c.init.data, "Include existing template data")
	flags.IntVarP(&c.init.depth, "depth", "d", c.init.depth, "Create a shallow clone")
	flags.VarP(c.init.exclude, "exclude", "x", "Exclude entry types")
	flags.BoolVar(&c.init.forcePromptOnce, "prompt", c.init.forcePromptOnce, "Force prompt*Once template functions to prompt") //nolint:lll
	flags.BoolVarP(&c.init.guessRepoURL, "guess-repo-url", "g", c.init.guessRepoURL, "Guess the repo URL")
	flags.BoolVar(&c.init.oneShot, "one-shot", c.init.oneShot, "Run in one-shot mode")
	flags.StringToStringVar(&c.init.promptBool, "promptBool", c.init.promptBool, "Populate promptBool")
	flags.StringToIntVar(&c.init.promptInt, "promptInt", c.init.promptInt, "Populate promptInt")
	flags.StringToStringVar(&c.init.promptString, "promptString", c.init.promptString, "Populate promptString")
	flags.BoolVarP(&c.init.purge, "purge", "p", c.init.purge, "Purge config and source directories after running")
	flags.BoolVarP(&c.init.purgeBinary, "purge-binary", "P", c.init.purgeBinary, "Purge chezmoi binary after running")
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

	// If we're not in a working tree then init it or clone it.
	gitDirAbsPath := c.WorkingTreeAbsPath.JoinString(git.GitDirName)
	switch fileInfo, err := c.baseSystem.Stat(gitDirAbsPath); {
	case err == nil && fileInfo.IsDir():
	case err == nil && !fileInfo.IsDir():
		return fmt.Errorf("%s: not a directory", gitDirAbsPath)
	case errors.Is(err, fs.ErrNotExist):
		workingTreeRawPath, err := c.baseSystem.RawPath(c.WorkingTreeAbsPath)
		if err != nil {
			return err
		}

		useBuiltinGit := c.UseBuiltinGit.Value(c.useBuiltinGitAutoFunc)

		if len(args) == 0 {
			if useBuiltinGit {
				if err := c.builtinGitInit(workingTreeRawPath); err != nil {
					return err
				}
			} else if err := c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"init", "--quiet"}); err != nil {
				return err
			}
		} else {
			var username, dotfilesRepoURL string
			if c.init.guessRepoURL {
				username, dotfilesRepoURL = guessDotfilesRepoURL(args[0], c.init.ssh)
			} else {
				dotfilesRepoURL = args[0]
			}
			if useBuiltinGit {
				if err := c.builtinGitClone(username, dotfilesRepoURL, workingTreeRawPath); err != nil {
					return err
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
					workingTreeRawPath.String(),
				)
				if err := c.run(chezmoi.EmptyAbsPath, c.Git.Command, args); err != nil {
					return err
				}
			}
		}
	case err != nil:
		return err
	}

	if err := c.checkVersion(); err != nil {
		return err
	}

	var err error
	c.SourceDirAbsPath, err = c.getSourceDirAbsPath(&getSourceDirAbsPathOptions{
		refresh: true,
	})
	if err != nil {
		return err
	}

	if err := c.createAndReloadConfigFile(); err != nil {
		return err
	}

	// Apply.
	if c.init.apply {
		if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, noArgs, applyArgsOptions{
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

// builtinGitClone clones a repo using the builtin git command.
func (c *Config) builtinGitClone(username, url string, workingTreeRawPath chezmoi.AbsPath) error {
	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return err
	}
	if c.init.ssh || endpoint.Protocol == "ssh" {
		return errors.New("builtin git does not support cloning repos over ssh, please install git")
	}

	isBare := false
	var referenceName plumbing.ReferenceName
	if c.init.branch != "" {
		referenceName = plumbing.NewBranchReferenceName(c.init.branch)
	}
	cloneOptions := git.CloneOptions{
		URL:               url,
		Depth:             c.init.depth,
		ReferenceName:     referenceName,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	for {
		_, err := git.PlainClone(workingTreeRawPath.String(), isBare, &cloneOptions)
		c.logger.Err(err).
			Stringer("path", workingTreeRawPath).
			Bool("isBare", isBare).
			Object("o", loggableGitCloneOptions(cloneOptions)).
			Msg("PlainClone")
		if !errors.Is(err, transport.ErrAuthenticationRequired) {
			return err
		}

		if _, err := fmt.Fprintf(c.stdout, "chezmoi: %s: %v\n", url, err); err != nil {
			return err
		}
		var basicAuth http.BasicAuth
		if basicAuth.Username, err = c.readString("Username? ", &username); err != nil {
			return err
		}
		if basicAuth.Username == "" {
			basicAuth.Username = username
		}
		if basicAuth.Password, err = c.readPassword("Password? "); err != nil {
			return err
		}
		cloneOptions.Auth = &basicAuth
	}
}

// builtinGitInit initializes a repo using the builtin git command.
func (c *Config) builtinGitInit(workingTreeRawPath chezmoi.AbsPath) error {
	isBare := false
	_, err := git.PlainInit(workingTreeRawPath.String(), isBare)
	c.logger.Err(err).
		Stringer("path", workingTreeRawPath).
		Bool("isBare", isBare).
		Msg("PlainInit")
	return err
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
//
// We cannot use zerolog's default object marshaler because it logs the auth
// credentials.
func (o loggableGitCloneOptions) MarshalZerologObject(e *zerolog.Event) {
	if o.URL != "" {
		e.Str("URL", o.URL)
	}
	if o.Auth != nil {
		e.Stringer("Auth", o.Auth)
	}
	if o.RemoteName != "" {
		e.Str("RemoteName", o.RemoteName)
	}
	if o.ReferenceName != "" {
		e.Stringer("ReferenceName", o.ReferenceName)
	}
	if o.SingleBranch {
		e.Bool("SingleBranch", o.SingleBranch)
	}
	if o.NoCheckout {
		e.Bool("NoCheckout", o.NoCheckout)
	}
	if o.Depth != 0 {
		e.Int("Depth", o.Depth)
	}
	if o.RecurseSubmodules != 0 {
		e.Uint("RecurseSubmodules", uint(o.RecurseSubmodules))
	}
	if o.Tags != 0 {
		e.Int("Tags", int(o.Tags))
	}
	if o.InsecureSkipTLS {
		e.Bool("InsecureSkipTLS", o.InsecureSkipTLS)
	}
	if o.CABundle != nil {
		e.Bytes("CABundle", o.CABundle)
	}
}

// guessDotfilesRepoURL guesses the user's username and dotfile repo from arg.
func guessDotfilesRepoURL(arg string, ssh bool) (username, repo string) {
	for _, dotfileRepoGuess := range dotfilesRepoGuesses {
		if !dotfileRepoGuess.rx.MatchString(arg) {
			continue
		}
		switch {
		case ssh && dotfileRepoGuess.sshRepoGuessRepl != "":
			repo = dotfileRepoGuess.rx.ReplaceAllString(arg, dotfileRepoGuess.sshRepoGuessRepl)
			return
		case !ssh && dotfileRepoGuess.httpRepoGuessRepl != "":
			username = dotfileRepoGuess.rx.ReplaceAllString(arg, dotfileRepoGuess.httpUsernameGuessRepl)
			repo = dotfileRepoGuess.rx.ReplaceAllString(arg, dotfileRepoGuess.httpRepoGuessRepl)
			return
		}
	}
	repo = arg
	return
}
