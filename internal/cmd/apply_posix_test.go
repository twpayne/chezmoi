// +build !windows

package cmd

import (
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/vfst"
)

func getApplyScriptTestCases(tempDir string) []scriptTestCase {
	return []scriptTestCase{
		{
			name: "simple",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true": "#!/bin/sh\necho foo >>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\nfoo\nfoo\n"),
				),
			},
		},
		{
			name: "simple_once",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_once_true": "#!/bin/sh\necho foo >>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\n"),
				),
			},
		},
		{
			name: "template",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true.tmpl": "#!/bin/sh\necho {{ .Foo }} >>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			data: map[string]interface{}{
				"Foo": "foo",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\nfoo\nfoo\n"),
				),
			},
		},
		{
			name: "issue_353",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"run_050_giraffe":       "#!/usr/bin/env bash\necho giraffe >>" + filepath.Join(tempDir, "evidence") + "\n",
					"run_150_elephant":      "#!/usr/bin/env bash\necho elephant >>" + filepath.Join(tempDir, "evidence") + "\n",
					"run_once_100_miauw.sh": "#!/usr/bin/env bash\necho miauw >>" + filepath.Join(tempDir, "evidence") + "\n",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString(strings.Join([]string{
						"giraffe\n",
						"miauw\n",
						"elephant\n",
						"giraffe\n",
						"elephant\n",
						"giraffe\n",
						"elephant\n",
					}, "")),
				),
			},
		},
	}
}

func getRunOnceFiles() map[string]interface{} {
	return map[string]interface{}{
		"/home/user/.local/share/chezmoi/run_once_foo.tmpl": "#!/bin/sh\necho bar >> {{ .TempFile }}\n",
	}
}
