package cmd

import "regexp"

type vcsInfo struct {
	cloneArgsFunc func(string, string) []string
	initArgs      []string
	pullArgs      []string
	versionArgs   []string
	versionRegexp *regexp.Regexp
}

var vcsInfos = map[string]*vcsInfo{
	"bzr": {
		versionArgs:   []string{"--version"},
		versionRegexp: regexp.MustCompile(`^Bazaar (bzr) (\d+\.\d+\.\d+)`),
	},
	"cvs": {
		versionArgs:   []string{"--version"},
		versionRegexp: regexp.MustCompile(`^Concurrent Versions System \(CVS\) (\d+\.\d+\.\d+)`),
	},
	"git": {
		cloneArgsFunc: func(repo, dir string) []string {
			return []string{"clone", repo, dir}
		},
		initArgs:      []string{"init"},
		pullArgs:      []string{"pull", "--rebase"},
		versionArgs:   []string{"version"},
		versionRegexp: regexp.MustCompile(`^git version (\d+\.\d+\.\d+)`),
	},
	"hg": {
		cloneArgsFunc: func(repo, dir string) []string {
			return []string{"clone", repo, dir}
		},
		initArgs:      []string{"init"},
		pullArgs:      []string{"pull", "--rebase", "--update"},
		versionArgs:   []string{"version"},
		versionRegexp: regexp.MustCompile(`^Mercurial Distributed SCM \(version (\d+\.\d+(\.\d+)?\))`),
	},
	"svn": {
		versionArgs:   []string{"--version"},
		versionRegexp: regexp.MustCompile(`^svn, version (\d+\.\d+\.\d+)`),
	},
}
