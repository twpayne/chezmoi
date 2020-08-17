package cmd

import (
	"os"
	"path/filepath"
	"regexp"
)

var hgVersionRegexp = regexp.MustCompile(`^Mercurial Distributed SCM \(version (\d+\.\d+(\.\d+)?\))`)

type hgVCS struct{}

func (hgVCS) AddArgs(path string) []string {
	return nil
}

func (hgVCS) CloneArgs(repo, dir string) []string {
	return []string{"clone", repo, dir}
}

func (hgVCS) CommitArgs(message string) []string {
	return nil
}

func (hgVCS) InitArgs() []string {
	return []string{"init"}
}

func (hgVCS) Initialized(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, ".hg"))
	switch {
	case err == nil:
		return info.IsDir(), nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}

func (hgVCS) ParseStatusOutput(output []byte) (interface{}, error) {
	return nil, nil
}

func (hgVCS) PullArgs() []string {
	return []string{"pull", "--update"}
}

func (hgVCS) PushArgs() []string {
	return nil
}

func (hgVCS) StatusArgs() []string {
	return nil
}

func (hgVCS) VersionArgs() []string {
	return []string{"version"}
}

func (hgVCS) VersionRegexp() *regexp.Regexp {
	return hgVersionRegexp
}
