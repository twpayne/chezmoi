package cmd

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/twpayne/chezmoi/internal/git"
)

var gitVersionRegexp = regexp.MustCompile(`^git version (\d+\.\d+\.\d+)`)

type gitVCS struct{}

func (gitVCS) AddArgs(path string) []string {
	return []string{"add", path}
}

func (gitVCS) CloneArgs(repo, dir string) []string {
	return []string{"clone", repo, dir}
}

func (gitVCS) CommitArgs(message string) []string {
	return []string{"commit", "--message", message}
}

func (gitVCS) InitArgs() []string {
	return []string{"init"}
}

func (gitVCS) Initialized(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	switch {
	case err == nil:
		return info.IsDir(), nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}

func (gitVCS) ParseStatusOutput(output []byte) (interface{}, error) {
	return git.ParseStatusPorcelainV2(output)
}

func (gitVCS) PullArgs() []string {
	return []string{"pull", "--rebase"}
}

func (gitVCS) PushArgs() []string {
	return []string{"push"}
}

func (gitVCS) StatusArgs() []string {
	return []string{"status", "--porcelain=v2"}
}

func (gitVCS) VersionArgs() []string {
	return []string{"version"}
}

func (gitVCS) VersionRegexp() *regexp.Regexp {
	return gitVersionRegexp
}
