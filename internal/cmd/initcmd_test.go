package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuessDotfilesRepoURL(t *testing.T) {
	for _, tc := range []struct {
		arg              string
		expectedHTTPSURL string
		expectedSSHURL   string
	}{
		{
			arg: "git@github.com:user/dotfiles.git",
		},
		{
			arg:              "gitlab.com/user",
			expectedHTTPSURL: "https://gitlab.com/user/dotfiles.git",
			expectedSSHURL:   "git@gitlab.com:user/dotfiles.git",
		},
		{
			arg:              "gitlab.com/user/dots",
			expectedHTTPSURL: "https://gitlab.com/user/dots.git",
			expectedSSHURL:   "git@gitlab.com:user/dots.git",
		},
		{
			arg:              "gitlab.com/user/dots.git",
			expectedHTTPSURL: "https://gitlab.com/user/dots.git",
			expectedSSHURL:   "git@gitlab.com:user/dots.git",
		},
		{
			arg:              "https://gitlab.com/user/dots.git",
			expectedHTTPSURL: "https://gitlab.com/user/dots.git",
		},
		{
			arg:              "sr.ht/~user",
			expectedHTTPSURL: "https://git.sr.ht/~user/dotfiles",
			expectedSSHURL:   "git@git.sr.ht:~user/dotfiles",
		},
		{
			arg:              "sr.ht/~user/dots",
			expectedHTTPSURL: "https://git.sr.ht/~user/dots",
			expectedSSHURL:   "git@git.sr.ht:~user/dots",
		},
		{
			arg:              "user",
			expectedHTTPSURL: "https://github.com/user/dotfiles.git",
			expectedSSHURL:   "git@github.com:user/dotfiles.git",
		},
		{
			arg:              "user/dots",
			expectedHTTPSURL: "https://github.com/user/dots.git",
			expectedSSHURL:   "git@github.com:user/dots.git",
		},
		{
			arg:              "user/dots.git",
			expectedHTTPSURL: "https://github.com/user/dots.git",
			expectedSSHURL:   "git@github.com:user/dots.git",
		},
	} {
		t.Run(tc.arg, func(t *testing.T) {
			expectedHTTPSURL := tc.expectedHTTPSURL
			if expectedHTTPSURL == "" {
				expectedHTTPSURL = tc.arg
			}
			assert.Equal(t, expectedHTTPSURL, guessDotfilesRepoURL(tc.arg, false))
			expectedSSHURL := tc.expectedSSHURL
			if tc.expectedSSHURL == "" {
				expectedSSHURL = tc.arg
			}
			assert.Equal(t, expectedSSHURL, guessDotfilesRepoURL(tc.arg, true))
		})
	}
}
