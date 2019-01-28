package cmd

import (
	"os"
	"testing"

	"github.com/twpayne/go-vfs/vfst"
)

// TestExercise exercises a few commands.
func TestExercise(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.bashrc": "# contents of .bashrc\n",
	})
	defer cleanup()
	if err != nil {
		t.Fatalf("vfst.NewTestFS() == _, _, %v, want _, _, <nil>", err)
	}

	c := &Config{
		SourceDir: "/home/user/.chezmoi",
		DestDir:   "/home/user",
		Umask:     022,
		Verbose:   true,
	}

	mustWriteFile := func(name, contents string, mode os.FileMode) {
		if err := fs.WriteFile(name, []byte(contents), mode); err != nil {
			t.Errorf("fs.WriteFile(%q, []byte(%q), %o) == %v, want <nil>", name, contents, mode, err)
		}
	}

	// chezmoi add ~/.bashrc
	t.Run("chezmoi_add_bashrc", func(t *testing.T) {
		if err := c.runAddCmd(fs, []string{"/home/user/.bashrc"}); err != nil {
			t.Errorf("c.runAddCmd(...) == %v, want <nil>", err)
		}
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
		if err := c.runForgetCmd(fs, []string{"/home/user/.bashrc"}); err != nil {
			t.Errorf("c.runForgetCmd(...) == %v, want <nil>", err)
		}
		vfst.RunTests(t, fs, "",
			vfst.TestPath("/home/user/.chezmoi/dot_bashrc",
				vfst.TestDoesNotExist,
			),
		)
	})

	// chezmoi add ~/.netrc
	t.Run("chezmoi_add_netrc", func(t *testing.T) {
		mustWriteFile("/home/user/.netrc", "# contents of .netrc\n", 0600)
		if err := c.runAddCmd(fs, []string{"/home/user/.netrc"}); err != nil {
			t.Errorf("c.runAddCmd(...) == %v, want <nil>", err)
		}
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
		if err := c.runChattrCmd(fs, []string{"-private,+empty", "/home/user/.netrc"}); err != nil {
			t.Errorf("c.runChattrCmd(...) == %v, want <nil>", err)
		}
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
		if err := c.runApplyCmd(fs, []string{"/home/user/.netrc"}); err != nil {
			t.Errorf("c.runApplyCmd(...) == %v, want <nil>", err)
		}
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
		if err := c.runRemoveCmd(fs, []string{"/home/user/.netrc"}); err != nil {
			t.Errorf("c.runRemoveCmd(...) == %v, want <nil>", err)
		}
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
