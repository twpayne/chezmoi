// +build windows

package cmd

import (
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/vfst"
)

type scriptTestCase struct {
	name  string
	root  interface{}
	data  map[string]interface{}
	tests []vfst.Test
}

func getApplyScriptTestCases(tempDir string) []scriptTestCase {
	return []scriptTestCase{
		{
			name: "simple",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true.#.bat": "@echo foo>>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\r\nfoo\r\nfoo\r\n"),
				),
			},
		},
		{
			name: "simple_once",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_once_true.#.bat": "@echo foo>>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\r\n"),
				),
			},
		},
		{
			name: "template",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true.#.bat.tmpl": "@echo {{ .Foo }}>>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			data: map[string]interface{}{
				"Foo": "foo",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\r\nfoo\r\nfoo\r\n"),
				),
			},
		},
		{
			name: "issue_353",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"run_050_giraffe":       "@echo giraffe>>" + filepath.Join(tempDir, "evidence") + "\n",
					"run_150_elephant":      "@echo elephant>>" + filepath.Join(tempDir, "evidence") + "\n",
					"run_once_100_miauw.sh": "@echo miauw>>" + filepath.Join(tempDir, "evidence") + "\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString(strings.Join([]string{
						"giraffe\r\n",
						"miauw\r\n",
						"elephant\r\n",
						"giraffe\r\n",
						"elephant\r\n",
						"giraffe\r\n",
						"elephant\r\n",
					}, "")),
				),
			},
		},
	}
}

func getRunOnceFiles() map[string]interface{} {
	return map[string]interface{}{
		"/home/user/.local/share/chezmoi/run_once_foo.#.bat.tmpl": "@echo bar>> {{ .TempFile }}\n",
	}
}
