package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	"github.com/twpayne/go-vfs/vfst"
	xdg "github.com/twpayne/go-xdg/v3"
)

func TestCreateConfigFile(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.local/share/chezmoi/.chezmoi.yaml.tmpl": strings.Join([]string{
			`{{ $email := promptString "email" -}}`,
			`data:`,
			`    email: "{{ $email }}"`,
			`    mailtoURL: "mailto:{{ $email }}"`,
			`    os: "{{ .chezmoi.os }}"`,
		}, "\n"),
	})
	require.NoError(t, err)
	defer cleanup()

	conf := &Config{
		SourceDir: "/home/user/.local/share/chezmoi",
		stdin:     bytes.NewBufferString("grace.hopper@example.com\n"),
		stdout:    &bytes.Buffer{},
		bds:       xdg.NewTestBaseDirectorySpecification("/home/user", nil),
	}

	require.NoError(t, conf.createConfigFile(fs, chezmoi.NewFSMutator(fs)))

	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.config/chezmoi/chezmoi.yaml",
			vfst.TestModeIsRegular,
			vfst.TestModePerm(0600),
			vfst.TestContentsString(strings.Join([]string{
				`data:`,
				`    email: "grace.hopper@example.com"`,
				`    mailtoURL: "mailto:grace.hopper@example.com"`,
				`    os: "` + runtime.GOOS + `"`,
			}, "\n")),
		),
	)

	assert.Equal(t, map[string]interface{}{
		"email":     "grace.hopper@example.com",
		"mailtourl": "mailto:grace.hopper@example.com",
		"os":        runtime.GOOS,
	}, conf.Data)
}
