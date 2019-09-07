package cmd

import "regexp"

// A VCS is a version control system.
type VCS interface {
	CloneArgs(string, string) []string
	InitArgs() []string
	PullArgs() []string
	VersionArgs() []string
	VersionRegexp() *regexp.Regexp
}

var vcses = map[string]VCS{
	"git": gitVCS{},
	"hg":  hgVCS{},
}
