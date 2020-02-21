package cmd

import "regexp"

// A VCS is a version control system.
type VCS interface {
	AddArgs(string) []string
	CloneArgs(string, string) []string
	CommitArgs(string) []string
	InitArgs() []string
	ParseStatusOutput([]byte) (interface{}, error)
	PullArgs() []string
	PushArgs() []string
	StatusArgs() []string
	VersionArgs() []string
	VersionRegexp() *regexp.Regexp
}

var vcses = map[string]VCS{
	"git": gitVCS{},
	"hg":  hgVCS{},
}
