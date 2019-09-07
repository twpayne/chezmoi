package cmd

import "regexp"

var hgVersionRegexp = regexp.MustCompile(`^Mercurial Distributed SCM \(version (\d+\.\d+(\.\d+)?\))`)

type hgVCS struct{}

func (hgVCS) CloneArgs(repo, dir string) []string {
	return []string{"clone", repo, dir}
}

func (hgVCS) InitArgs() []string {
	return []string{"init"}
}

func (hgVCS) PullArgs() []string {
	return []string{"pull", "--rebase", "--update"}
}

func (hgVCS) VersionArgs() []string {
	return []string{"version"}
}

func (hgVCS) VersionRegexp() *regexp.Regexp {
	return hgVersionRegexp
}
