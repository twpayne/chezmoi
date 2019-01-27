package cmd

import (
	"regexp"
)

type vcsInfo struct {
	cloneArgsFunc func(string, string) []string
	pullArgs      []string
	versionArgs   []string
	versionRegexp *regexp.Regexp
}

var (
	vcsInfos = map[string]*vcsInfo{
		"bzr": &vcsInfo{
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^Bazaar (bzr) (\d+\.\d+\.\d+)`),
		},
		"cvs": &vcsInfo{
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^Concurrent Versions System \(CVS\) (\d+\.\d+\.\d+)`),
		},
		"git": &vcsInfo{
			cloneArgsFunc: func(repo, dir string) []string {
				return []string{"clone", repo, dir}
			},
			pullArgs:      []string{"pull", "--rebase"},
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`^git version (\d+\.\d+\.\d+)`),
		},
		"hg": &vcsInfo{
			cloneArgsFunc: func(repo, dir string) []string {
				return []string{"clone", repo, dir}
			},
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`^Mercurial Distributed SCM \(version (\d+\.\d+\.\d+\))`),
		},
		"svn": &vcsInfo{
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^svn, version (\d+\.\d+\.\d+)`),
		},
	}
)
