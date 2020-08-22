package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"

	"github.com/twpayne/chezmoi/cmd"
)

//nolint:interfacer
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"chezmoi": func() int {
			if err := cmd.Execute(); err != nil {
				if s := err.Error(); s != "" {
					fmt.Fprintf(os.Stderr, "chezmoi: %s\n", s)
				}
				return 1
			}
			return 0
		},
	}))
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "scripts"),
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"chhome":      cmdChHome,
			"cmpmod":      cmdCmpMod,
			"edit":        cmdEdit,
			"mkfile":      cmdMkFile,
			"mkhomedir":   cmdMkHomeDir,
			"mksourcedir": cmdMkSourceDir,
		},
		Condition: func(cond string) (bool, error) {
			switch cond {
			case "windows":
				return runtime.GOOS == "windows", nil
			default:
				return false, fmt.Errorf("%s: unknown condition", cond)
			}
		},
		Setup:         setup,
		UpdateScripts: os.Getenv("CHEZMOIUPDATESCRIPTS") != "",
	})
}

// cmdChHome changes the home directory to its argument, creating the directory
// if it does not already exists. It updates the HOME environment variable, and,
// if running on Windows, USERPROFILE too.
func cmdChHome(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! chhome")
	}
	if len(args) != 1 {
		ts.Fatalf("usage: chhome dir")
	}
	var (
		homeDir          = ts.MkAbs(args[0])
		chezmoiConfigDir = filepath.Join(homeDir, ".config", "chezmoi")
		chezmoiSourceDir = filepath.Join(homeDir, ".local", "share", "chezmoi")
	)
	ts.Check(os.MkdirAll(homeDir, 0o777))
	ts.Setenv("HOME", homeDir)
	ts.Setenv("CHEZMOICONFIGDIR", chezmoiConfigDir)
	ts.Setenv("CHEZMOISOURCEDIR", chezmoiSourceDir)
	if runtime.GOOS == "windows" {
		ts.Setenv("USERPROFILE", homeDir)
	}
}

// cmdCmpMod compares modes.
func cmdCmpMod(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: cmpmod mode path")
	}
	mode64, err := strconv.ParseUint(args[0], 8, 32)
	if err != nil || os.FileMode(mode64)&os.ModePerm != os.FileMode(mode64) {
		ts.Fatalf("invalid mode: %s", args[0])
	}
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(args[1])
	if err != nil {
		ts.Fatalf("%s: %v", args[1], err)
	}
	equal := info.Mode()&os.ModePerm == os.FileMode(mode64)
	if neg && equal {
		ts.Fatalf("%s unexpectedly has mode %03o", args[1], info.Mode()&os.ModePerm)
	}
	if !neg && !equal {
		ts.Fatalf("%s has mode %03o, expected %03o", args[1], info.Mode()&os.ModePerm, os.FileMode(mode64))
	}
}

// cmdEdit edits all of its arguments by appending "# edited\n" to them.
func cmdEdit(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! edit")
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			ts.Fatalf("edit: %v", err)
		}
		data = append(data, []byte("# edited\n")...)
		//nolint:gosec
		if err := ioutil.WriteFile(filename, data, 0o666); err != nil {
			ts.Fatalf("edit: %v", err)
		}
	}
}

// cmdMkFile creates empty files.
func cmdMkFile(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkfile")
	}
	perm := os.FileMode(0o666)
	if len(args) >= 1 && strings.HasPrefix(args[0], "-perm=") {
		permStr := strings.TrimPrefix(args[0], "-perm=")
		permUint32, err := strconv.ParseUint(permStr, 8, 32)
		if err != nil {
			ts.Fatalf("%s: bad permissions", permStr)
		}
		perm = os.FileMode(permUint32)
		args = args[1:]
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		_, err := os.Lstat(filename)
		switch {
		case err == nil:
			ts.Fatalf("%s: already exists", arg)
		case !os.IsNotExist(err):
			ts.Fatalf("%s: %v", arg, err)
		}
		if err := ioutil.WriteFile(filename, nil, perm); err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
	}
}

// cmdMkHomeDir makes and populates a home directory.
func cmdMkHomeDir(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkhomedir")
	}
	if len(args) > 1 {
		ts.Fatalf(("usage: mkhomedir [path]"))
	}
	path := ts.Getenv("HOME")
	if len(args) > 0 {
		path = ts.MkAbs(args[0])
	}
	workDir := ts.Getenv("WORK")
	relPath, err := filepath.Rel(workDir, path)
	ts.Check(err)
	if err := vfst.NewBuilder().Build(vfs.NewPathFS(vfs.OSFS, workDir), map[string]interface{}{
		relPath: map[string]interface{}{
			".bashrc": "# contents of .bashrc\n",
			".binary": &vfst.File{
				Perm:     0o755,
				Contents: []byte("#!/bin/sh\n"),
			},
			".exists": "# contents of .exists\n",
			".gitconfig": "" +
				"[user]\n" +
				"  email = you@example.com\n" +
				"  name = Your Name\n",
			".hushlogin": "",
			".ssh": &vfst.Dir{
				Perm: 0o700,
				Entries: map[string]interface{}{
					"config": "# contents of .ssh/config\n",
				},
			},
			".symlink": &vfst.Symlink{
				Target: ".bashrc",
			},
		},
	}); err != nil {
		ts.Fatalf("mkhomedir: %v", err)
	}
}

// cmdMkSourceDir makes and populates a source directory.
func cmdMkSourceDir(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mksourcedir")
	}
	if len(args) > 1 {
		ts.Fatalf("usage: mksourcedir [path]")
	}
	sourceDir := ts.Getenv("CHEZMOISOURCEDIR")
	if len(args) > 0 {
		sourceDir = ts.MkAbs(args[0])
	}
	workDir := ts.Getenv("WORK")
	relPath, err := filepath.Rel(workDir, sourceDir)
	ts.Check(err)
	err = vfst.NewBuilder().Build(vfs.NewPathFS(vfs.OSFS, workDir), map[string]interface{}{
		relPath: map[string]interface{}{
			"dot_absent":            "",
			"empty_dot_hushlogin":   "",
			"executable_dot_binary": "#!/bin/sh\n",
			"dot_bashrc":            "# contents of .bashrc\n",
			"dot_gitconfig.tmpl": "" +
				"[user]\n" +
				"  email = {{ \"you@example.com\" }}\n" +
				"  name = Your Name\n",
			"private_dot_ssh": map[string]interface{}{
				"config": "# contents of .ssh/config\n",
			},
			"symlink_dot_symlink": ".bashrc\n",
		},
	})
	if err != nil {
		ts.Fatalf("mksourcedir: %v", err)
	}
}

func setup(env *testscript.Env) error {
	var (
		binDir           = filepath.Join(env.WorkDir, "bin")
		homeDir          = filepath.Join(env.WorkDir, "home", "user")
		chezmoiConfigDir = filepath.Join(homeDir, ".config", "chezmoi")
		chezmoiSourceDir = filepath.Join(homeDir, ".local", "share", "chezmoi")
	)

	env.Setenv("HOME", homeDir)
	env.Setenv("PATH", prependDirToPath(binDir, env.Getenv("PATH")))
	env.Setenv("CHEZMOICONFIGDIR", chezmoiConfigDir)
	env.Setenv("CHEZMOISOURCEDIR", chezmoiSourceDir)
	switch runtime.GOOS {
	case "windows":
		env.Setenv("EDITOR", filepath.Join(binDir, "editor.cmd"))
		env.Setenv("USERPROFILE", homeDir)
		// There is not currently a convenient way to override the shell on
		// Windows.
	default:
		env.Setenv("EDITOR", filepath.Join(binDir, "editor"))
		env.Setenv("SHELL", filepath.Join(binDir, "shell"))
	}

	if runtime.GOOS != "windows" {
		// Fix permissions on any files in the bin directory.
		infos, err := ioutil.ReadDir(binDir)
		if err == nil {
			for _, info := range infos {
				if err := os.Chmod(filepath.Join(binDir, info.Name()), 0o755); err != nil {
					return err
				}
			}
		}
	}

	root := make(map[string]interface{})
	switch runtime.GOOS {
	case "windows":
		root["/bin"] = map[string]interface{}{
			// editor.cmd a non-interactive script that appends "# edited\n" to
			// the end of each file.
			"editor.cmd": &vfst.File{
				Perm:     0o755,
				Contents: []byte(`@for %%x in (%*) do echo # edited >> %%x`),
			},
		}
	default:
		root["/bin"] = map[string]interface{}{
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
		}
	}

	return vfst.NewBuilder().Build(vfs.NewPathFS(vfs.OSFS, env.WorkDir), root)
}

func prependDirToPath(dir, path string) string {
	return strings.Join(append([]string{dir}, filepath.SplitList(path)...), string(os.PathListSeparator))
}
