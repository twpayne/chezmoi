package cmd

import (
	"io/fs"
	"runtime"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

var _ fs.FileInfo = &fileInfo{}

func TestIssue3891(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}

	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user": map[string]any{
			"run.sh": "#!/bin/sh\n",
			".local/share/chezmoi": map[string]any{
				"executable_run.sh": "#!/bin/sh",
			},
		},
	}, func(fileSystem vfs.FS) {
		assert.NoError(t, newTestConfig(t, fileSystem).execute([]string{"re-add"}))
		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath("/home/user/.local/share/chezmoi/executable_run.sh",
				vfst.TestContentsString("#!/bin/sh\n"),
			),
			vfst.TestPath("/home/user/.local/share/chezmoi/run.sh",
				vfst.TestDoesNotExist(),
			),
		)
	})
}
