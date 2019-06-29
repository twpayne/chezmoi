package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

// TestExercise exercises a few commands.
func TestExercise(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.bashrc": "# contents of .bashrc\n",
	})
	require.NoError(t, err)
	defer cleanup()

	c := &Config{
		SourceDir: "/home/user/.chezmoi",
		DestDir:   "/home/user",
		Umask:     022,
		Verbose:   true,
		remove: removeCmdConfig{
			force: true,
		},
	}

	mustWriteFile := func(name, contents string, mode os.FileMode) {
		assert.NoError(t, fs.WriteFile(name, []byte(contents), mode))
	}

	// chezmoi add ~/.bashrc
	t.Run("chezmoi_add_bashrc", func(t *testing.T) {
		assert.NoError(t, c.runAddCmd(fs, []string{"/home/user/.bashrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.chezmoi",
				vfst.TestIsDir,
				vfst.TestModePerm(0700),
			),
			vfst.TestPath("/home/user/.chezmoi/dot_bashrc",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0644),
				vfst.TestContentsString("# contents of .bashrc\n"),
			),
		)
	})

	// chezmoi forget ~/.bashrc
	t.Run("chezmoi_forget_bashrc", func(t *testing.T) {
		assert.NoError(t, c.runForgetCmd(fs, []string{"/home/user/.bashrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.chezmoi/dot_bashrc",
				vfst.TestDoesNotExist,
			),
		)
	})

	// chezmoi add ~/.netrc
	t.Run("chezmoi_add_netrc", func(t *testing.T) {
		mustWriteFile("/home/user/.netrc", "# contents of .netrc\n", 0600)
		assert.NoError(t, c.runAddCmd(fs, []string{"/home/user/.netrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.chezmoi/private_dot_netrc",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0644),
				vfst.TestContentsString("# contents of .netrc\n"),
			),
		)
	})

	// chezmoi chattr -- -private,+empty ~/.netrc
	t.Run("chezmoi_chattr_netrc", func(t *testing.T) {
		assert.NoError(t, c.runChattrCmd(fs, []string{"-private,+empty", "/home/user/.netrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.chezmoi/empty_dot_netrc",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0644),
				vfst.TestContentsString("# contents of .netrc\n"),
			),
		)
	})

	// chezmoi apply ~/.netrc
	t.Run("chezmoi_apply_netrc", func(t *testing.T) {
		assert.NoError(t, c.runApplyCmd(fs, []string{"/home/user/.netrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.netrc",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0644),
				vfst.TestContentsString("# contents of .netrc\n"),
			),
		)
	})

	// chezmoi remove ~/.netrc
	t.Run("chezmoi_remove_netrc", func(t *testing.T) {
		assert.NoError(t, c.runRemoveCmd(fs, []string{"/home/user/.netrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.netrc",
				vfst.TestDoesNotExist,
			),
			vfst.TestPath("/home/user/.chezmoi/empty_dot_netrc",
				vfst.TestDoesNotExist,
			),
		)
	})
}
