package cmd

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestGuessRepoURL(t *testing.T) {
	for _, tc := range []struct {
		arg                 string
		expectedHTTPRepoURL string
		expectedSSHRepoURL  string
		expectedUsername    string
	}{
		{
			arg:                 "git@github.com:user/dotfiles.git",
			expectedHTTPRepoURL: "git@github.com:user/dotfiles.git",
			expectedSSHRepoURL:  "git@github.com:user/dotfiles.git",
		},
		{
			arg:                 "codeberg.org/user",
			expectedHTTPRepoURL: "https://codeberg.org/user/dotfiles.git",
			expectedSSHRepoURL:  "git@codeberg.org:user/dotfiles.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "codeberg.org/user/dots",
			expectedHTTPRepoURL: "https://codeberg.org/user/dots.git",
			expectedSSHRepoURL:  "git@codeberg.org:user/dots.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "gitlab.com/user",
			expectedHTTPRepoURL: "https://gitlab.com/user/dotfiles.git",
			expectedSSHRepoURL:  "git@gitlab.com:user/dotfiles.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "gitlab.com/user/dots",
			expectedHTTPRepoURL: "https://gitlab.com/user/dots.git",
			expectedSSHRepoURL:  "git@gitlab.com:user/dots.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "gitlab.com/user/dots.git",
			expectedHTTPRepoURL: "https://gitlab.com/user/dots.git",
			expectedSSHRepoURL:  "git@gitlab.com:user/dots.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "http://gitlab.com/user/dots.git",
			expectedHTTPRepoURL: "http://gitlab.com/user/dots.git",
			expectedSSHRepoURL:  "git@gitlab.com:user/dots.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "https://gitlab.com/user/dots.git",
			expectedHTTPRepoURL: "https://gitlab.com/user/dots.git",
			expectedSSHRepoURL:  "git@gitlab.com:user/dots.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "sr.ht/~user_name",
			expectedHTTPRepoURL: "https://git.sr.ht/~user_name/dotfiles",
			expectedSSHRepoURL:  "git@git.sr.ht:~user_name/dotfiles",
			expectedUsername:    "user_name",
		},
		{
			arg:                 "sr.ht/~user_name/dots",
			expectedHTTPRepoURL: "https://git.sr.ht/~user_name/dots",
			expectedSSHRepoURL:  "git@git.sr.ht:~user_name/dots",
			expectedUsername:    "user_name",
		},
		{
			arg:                 "user",
			expectedHTTPRepoURL: "https://github.com/user/dotfiles.git",
			expectedSSHRepoURL:  "git@github.com:user/dotfiles.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "user/dots",
			expectedHTTPRepoURL: "https://github.com/user/dots.git",
			expectedSSHRepoURL:  "git@github.com:user/dots.git",
			expectedUsername:    "user",
		},
		{
			arg:                 "user/dots.git",
			expectedHTTPRepoURL: "https://github.com/user/dots.git",
			expectedSSHRepoURL:  "git@github.com:user/dots.git",
			expectedUsername:    "user",
		},
	} {
		t.Run(tc.arg, func(t *testing.T) {
			ssh := false
			actualHTTPUsername, actualHTTPRepoURL := guessRepoURL(tc.arg, ssh)
			assert.Equal(t, tc.expectedUsername, actualHTTPUsername, "HTTPUsername")
			assert.Equal(t, tc.expectedHTTPRepoURL, actualHTTPRepoURL, "HTTPRepoURL")

			ssh = true
			actualSSHUsername, actualSSHRepoURL := guessRepoURL(tc.arg, ssh)
			assert.Equal(t, "", actualSSHUsername, "SSHUsername")
			assert.Equal(t, tc.expectedSSHRepoURL, actualSSHRepoURL, "SSHRepoURL")
		})
	}
}

func TestIssue2137(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user/.local/share/chezmoi": map[string]any{
			".chezmoiversion": "3.0.0",
			".git": map[string]any{
				".keep": nil,
			},
		},
	}, func(fileSystem vfs.FS) {
		err := newTestConfig(t, fileSystem).execute([]string{"init"})
		tooOldError := &chezmoi.TooOldError{}
		require.ErrorAs(t, err, &tooOldError)
	})
}

func TestIssue2283(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping UNIX test on Windows")
	}
	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user/.local/share/chezmoi": map[string]any{
			".chezmoiroot": "home",
			"home": map[string]any{
				".chezmoi.yaml.tmpl": "sourceDir: {{ .chezmoi.sourceDir }}\n",
			},
		},
	}, func(fileSystem vfs.FS) {
		require.NoError(t, newTestConfig(t, fileSystem).execute([]string{"init"}))
		data, err := fileSystem.ReadFile("/home/user/.config/chezmoi/chezmoi.yaml")
		require.NoError(t, err)
		assert.Equal(t, "sourceDir: /home/user/.local/share/chezmoi/home\n", string(data))
	})
}
