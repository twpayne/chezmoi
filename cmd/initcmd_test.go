package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuessDotfilesRepoURL(t *testing.T) {
	for argStr, expected := range map[string]string{
		"git@github.com:user/dotfiles.git": "git@github.com:user/dotfiles.git",
		"gitlab.com/user":                  "https://gitlab.com/user/dotfiles.git",
		"gitlab.com/user/dots":             "https://gitlab.com/user/dots.git",
		"gitlab.com/user/dots.git":         "https://gitlab.com/user/dots.git",
		"https://gitlab.com/user/dots.git": "https://gitlab.com/user/dots.git",
		"sr.ht/~user":                      "https://git.sr.ht/~user/dotfiles",
		"sr.ht/~user/dots":                 "https://git.sr.ht/~user/dots",
		"user":                             "https://github.com/user/dotfiles.git",
		"user/dots":                        "https://github.com/user/dots.git",
		"user/dots.git":                    "https://github.com/user/dots.git",
	} {
		assert.Equal(t, expected, guessDotfilesRepoURL(argStr))
	}
}
