package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"
)

//nolint:interfacer
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"chezmoi": testRun,
	}))
}

func TestChezmoi(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "scripts"),
		Condition: func(cond string) (bool, error) {
			switch cond {
			case "windows":
				return runtime.GOOS == "windows", nil
			default:
				return false, fmt.Errorf("unknown condition: %s", cond)
			}
		},
		Setup: func(env *testscript.Env) error {
			switch runtime.GOOS {
			case "windows":
				return setupWindowsEnv(env)
			default:
				return setupPOSIXEnv(env)
			}
		},
	})
}

func testRun() int {
	if err := run(); err != nil {
		if s := err.Error(); s != "" {
			fmt.Printf("chezmoi: %s\n", s)
		}
		return 1
	}
	return 0
}

func setupPOSIXEnv(env *testscript.Env) error {
	binDir := filepath.Join(env.WorkDir, "bin")
	env.Setenv("EDITOR", filepath.Join(binDir, "editor"))
	env.Setenv("HOME", filepath.Join(env.WorkDir, "home", "user"))
	env.Setenv("PATH", prependDirToPath(binDir, env.Getenv("PATH")))
	env.Setenv("SHELL", filepath.Join(binDir, "shell"))

	return vfst.NewBuilder().Build(vfs.NewPathFS(vfs.HostOSFS, env.WorkDir), map[string]interface{}{
		"/bin": map[string]interface{}{
			// editor a non-interactive script that appends "# edited\n" to the
			// end of each file.
			"editor": &vfst.File{
				Perm: 0o755,
				Contents: []byte(strings.Join([]string{
					`#!/bin/sh`,
					``,
					`for filename in $*; do`,
					`    echo "# edited" >> $filename`,
					`done`,
				}, "\n")),
			},
			// shell is a non-interactive script that appends the directory in
			// which it was launched to $WORK/shell.log.
			"shell": &vfst.File{
				Perm: 0o755,
				Contents: []byte(strings.Join([]string{
					`#!/bin/sh`,
					``,
					`echo $PWD >> ` + filepath.Join(env.WorkDir, "shell.log"),
				}, "\n")),
			},
		},
		"/home/user": map[string]interface{}{
			// .gitconfig is populated with a user and email to avoid warnings
			// from git.
			".gitconfig": strings.Join([]string{
				`[user]`,
				`    name = Username`,
				`    email = user@home.org`,
			}, "\n"),
		},
	})
}

func setupWindowsEnv(env *testscript.Env) error {
	// FIXME
	return nil
}

func prependDirToPath(dir, path string) string {
	return strings.Join(append([]string{dir}, filepath.SplitList(path)...), string(os.PathListSeparator))
}
