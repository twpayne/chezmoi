package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Masterminds/sprig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/go-vfs/vfst"
	xdg "github.com/twpayne/go-xdg/v3"
)

func TestCreateConfigFile(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.local/share/chezmoi/.chezmoi.yaml.tmpl": strings.Join([]string{
			`{{ $email := promptString "email" | trim -}}`,
			`data:`,
			`    email: "{{ $email }}"`,
			`    mailtoURL: "mailto:{{ $email }}"`,
			`    os: "{{ .chezmoi.os }}"`,
		}, "\n"),
	})
	require.NoError(t, err)
	defer cleanup()

	conf := &Config{
		fs:            fs,
		mutator:       chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false, 0),
		SourceDir:     "/home/user/.local/share/chezmoi",
		templateFuncs: sprig.HermeticTxtFuncMap(),
		stdin:         bytes.NewBufferString("john.smith@company.com \n"),
		stdout:        &bytes.Buffer{},
		bds:           xdg.NewTestBaseDirectorySpecification("/home/user", nil),
	}

	require.NoError(t, conf.createConfigFile())

	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.config/chezmoi/chezmoi.yaml",
			vfst.TestModeIsRegular,
			vfst.TestModePerm(0600),
			vfst.TestContentsString(strings.Join([]string{
				`data:`,
				`    email: "john.smith@company.com"`,
				`    mailtoURL: "mailto:john.smith@company.com"`,
				`    os: "` + runtime.GOOS + `"`,
			}, "\n")),
		),
	)

	assert.Equal(t, map[string]interface{}{
		"email":     "john.smith@company.com",
		"mailtourl": "mailto:john.smith@company.com",
		"os":        runtime.GOOS,
	}, conf.Data)
}

func TestInit(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user": &vfst.Dir{Perm: 0755},
	})
	require.NoError(t, err)
	defer cleanup()

	c := &Config{
		fs:        fs,
		mutator:   chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false, 0),
		SourceDir: "/home/user/.local/share/chezmoi",
		SourceVCS: sourceVCSConfig{
			Command: "git",
		},
		bds: xdg.NewTestBaseDirectorySpecification("/home/user", func(string) string { return "" }),
	}

	require.NoError(t, c.runInitCmd(nil, nil))
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.local/share/chezmoi",
			vfst.TestIsDir,
		),
		vfst.TestPath("/home/user/.local/share/chezmoi/.git",
			vfst.TestIsDir,
		),
		vfst.TestPath("/home/user/.local/share/chezmoi/.git/HEAD",
			vfst.TestModeIsRegular,
		),
	)
}

func TestInitRepo(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user": &vfst.Dir{Perm: 0755},
	})
	require.NoError(t, err)
	defer cleanup()

	c := &Config{
		fs:        fs,
		mutator:   chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false, 0),
		SourceDir: "/home/user/.local/share/chezmoi",
		SourceVCS: sourceVCSConfig{
			Command: "git",
		},
		bds: xdg.NewTestBaseDirectorySpecification("/home/user", func(string) string { return "" }),
	}

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, c.runInitCmd(nil, []string{filepath.Join(wd, "testdata/gitrepo")}))
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.local/share/chezmoi",
			vfst.TestIsDir,
		),
		vfst.TestPath("/home/user/.local/share/chezmoi/.git",
			vfst.TestIsDir,
		),
		vfst.TestPath("/home/user/.local/share/chezmoi/.git/HEAD",
			vfst.TestModeIsRegular,
		),
		vfst.TestPath("/home/user/.local/share/chezmoi/dot_bashrc",
			vfst.TestModeIsRegular,
			vfst.TestContentsString(lines("# contents of .bashrc\n")),
		),
		vfst.TestPath("/home/user/.config/chezmoi/chezmoi.toml",
			vfst.TestModeIsRegular,
			vfst.TestContentsString(lines("# contents of chezmoi.toml\n")),
		),
	)
}
