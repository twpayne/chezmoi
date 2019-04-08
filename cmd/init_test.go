package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
	xdg "github.com/twpayne/go-xdg/v3"
)

func TestCreateConfigFile(t *testing.T) {
	path := "/home/user/.local/share/chezmoi/.chezmoi.yaml.tmpl"
	content := `
{{- $email := promptString "email" }}
data:
    email: "{{ $email }}"
    mailtoURL: "mailto:{{ $email }}"
`

	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{path: content}, vfst.BuilderVerbose(true))
	require.NoError(t, err)
	defer cleanup()

	out := testingWriter{t}
	in := new(bytes.Buffer)
	in.WriteString("grace.hopper@example.com\n")

	conf := &Config{
		SourceDir: "/home/user/.local/share/chezmoi",
		stdout:    out,
		stdin:     in,
		bds:       &xdg.BaseDirectorySpecification{ConfigHome: "/home/user/.config"},
	}

	err = conf.createConfigFile(fs)
	require.NoError(t, err)

	expected := `
data:
    email: "grace.hopper@example.com"
    mailtoURL: "mailto:grace.hopper@example.com"
`

	expectedPath := "/home/user/.config/chezmoi/chezmoi.yaml"
	actual, err := fs.ReadFile(expectedPath)
	if os.IsNotExist(err) {
		t.Fatalf("Configuration file was not written to %q", expectedPath)
	}

	require.NoError(t, err)
	assert.Equal(t, expected, string(actual))

	s, err := fs.Stat(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, "-rw-r--r--", s.Mode().String(), "bad file mode")

	// Make sure we are loading the config that we just created
	assert.Equal(t, map[string]interface{}{
		"email":     "grace.hopper@example.com",
		"mailtourl": "mailto:grace.hopper@example.com",
	}, conf.Data)
}

// The testingWriter implements io.Writer by sending all written bytes to the
// testing.T log.
type testingWriter struct {
	*testing.T
}

func (w testingWriter) Write(data []byte) (int, error) {
	msg := string(data)
	if msg[len(msg)-1] == '\n' {
		// The Log function of the testing.T automatically adds a newline
		msg = msg[:len(msg)-1]
	}

	w.Log(msg)
	return len(data), nil
}
