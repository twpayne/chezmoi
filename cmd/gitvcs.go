package cmd

import "regexp"

var gitVersionRegexp = regexp.MustCompile(`^git version (\d+\.\d+\.\d+)`)

type gitVCS struct{}

func (gitVCS) CloneArgs(repo, dir string) []string {
	return []string{"clone", repo, dir}
}

func (gitVCS) InitArgs() []string {
	return []string{"init"}
}

func (gitVCS) PullArgs() []string {
	return []string{"pull", "--rebase"}
}

func (gitVCS) VersionArgs() []string {
	return []string{"version"}
}

func (gitVCS) VersionRegexp() *regexp.Regexp {
	return gitVersionRegexp
}
