package cmd

import (
	"os"
	"runtime"
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

	c := newTestConfig(
		fs,
		withRemoveCmdConfig(removeCmdConfig{
			force: true,
		}),
	)

	mustWriteFile := func(name, contents string, mode os.FileMode) {
		require.NoError(t, fs.WriteFile(name, []byte(contents), mode))
	}

	// chezmoi add ~/.bashrc
	t.Run("chezmoi_add_bashrc", func(t *testing.T) {
		assert.NoError(t, c.runAddCmd(nil, []string{"/home/user/.bashrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.local/share/chezmoi",
				vfst.TestIsDir,
				vfst.TestModePerm(0o700),
			),
			vfst.TestPath("/home/user/.local/share/chezmoi/dot_bashrc",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0o644),
				vfst.TestContentsString("# contents of .bashrc\n"),
			),
		)
	})

	// chezmoi forget ~/.bashrc
	t.Run("chezmoi_forget_bashrc", func(t *testing.T) {
		assert.NoError(t, c.runForgetCmd(nil, []string{"/home/user/.bashrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.local/share/chezmoi/dot_bashrc",
				vfst.TestDoesNotExist,
			),
		)
	})

	// chezmoi add ~/.netrc
	t.Run("chezmoi_add_netrc", func(t *testing.T) {
		mustWriteFile("/home/user/.netrc", "# contents of .netrc\n", 0o600)
		assert.NoError(t, c.runAddCmd(nil, []string{"/home/user/.netrc"}))
		path := "/home/user/.local/share/chezmoi/private_dot_netrc"
		// Private files are not supported on Windows.
		if runtime.GOOS == "windows" {
			path = "/home/user/.local/share/chezmoi/dot_netrc"
		}
		vfst.RunTests(t, fs, "",
			vfst.TestPath(path,
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0o644),
				vfst.TestContentsString("# contents of .netrc\n"),
			),
		)
	})

	// chezmoi chattr -- -private,+empty ~/.netrc
	t.Run("chezmoi_chattr_netrc", func(t *testing.T) {
		assert.NoError(t, c.runChattrCmd(nil, []string{"-private,+empty", "/home/user/.netrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.local/share/chezmoi/empty_dot_netrc",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0o644),
				vfst.TestContentsString("# contents of .netrc\n"),
			),
		)
	})

	// chezmoi apply ~/.netrc
	t.Run("chezmoi_apply_netrc", func(t *testing.T) {
		assert.NoError(t, c.runApplyCmd(nil, []string{"/home/user/.netrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.netrc",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0o644),
				vfst.TestContentsString("# contents of .netrc\n"),
			),
		)
	})

	// chezmoi remove ~/.netrc
	t.Run("chezmoi_remove_netrc", func(t *testing.T) {
		assert.NoError(t, c.runRemoveCmd(nil, []string{"/home/user/.netrc"}))
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.netrc",
				vfst.TestDoesNotExist,
			),
			vfst.TestPath("/home/user/.local/share/chezmoi/empty_dot_netrc",
				vfst.TestDoesNotExist,
			),
		)
	})
}
