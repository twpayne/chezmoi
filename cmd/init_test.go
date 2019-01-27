package cmd

import (
	"testing"

	"github.com/twpayne/go-vfs/vfst"
)

func TestInit(t *testing.T) {
	c := &Config{
		SourceDir:        "/home/user/.local/share/chezmoi",
		DestDir:          "/home/user",
		SourceVCSCommand: "test", // hack to work around untestable Config.exec(…)
		Umask:            022,
		DryRun:           false,
		Verbose:          true,
	}
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user":          &vfst.Dir{Perm: 0755},
		"/home/user/.chezmoi": &vfst.Dir{Perm: 0700},
		"/home/user/.bashrc":  "# contents of .bashrc\n",
	})
	defer cleanup()
	if err != nil {
		t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
	}
	args := []string{"git@github.com:example/dotfiles.git"}
	if err := c.runInitCommand(fs, args); err != nil {
		t.Errorf("c.runInitCommand(fs, nil, %+v) == %v, want <nil>", args, err)
	}
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.local",
			vfst.TestIsDir,
			vfst.TestModePerm(0700),
		),
		vfst.TestPath("/home/user/.local/share",
			vfst.TestIsDir,
			vfst.TestModePerm(0700),
		),
		vfst.TestPath("/home/user/.local/share/chezmoi",
			vfst.TestIsDir,
			vfst.TestModePerm(0700),
		),
	)
}
