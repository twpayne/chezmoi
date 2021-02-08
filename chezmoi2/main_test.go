package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"

	"github.com/twpayne/chezmoi/chezmoi2/cmd"
	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoitest"
)

//nolint:interfacer
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"chezmoi": func() int {
			return cmd.Main(cmd.VersionInfo{
				Version: "v2.0.0+test",
				Commit:  "HEAD",
				Date:    time.Now().UTC().Format(time.RFC3339),
				BuiltBy: "testscript",
			}, os.Args[1:])
		},
	}))
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "scripts"),
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"chhome":         cmdChHome,
			"cmpmod":         cmdCmpMod,
			"edit":           cmdEdit,
			"mkfile":         cmdMkFile,
			"mkageconfig":    cmdMkAGEConfig,
			"mkgitconfig":    cmdMkGitConfig,
			"mkgpgconfig":    cmdMkGPGConfig,
			"mkhomedir":      cmdMkHomeDir,
			"mksourcedir":    cmdMkSourceDir,
			"rmfinalnewline": cmdRmFinalNewline,
			"unix2dos":       cmdUNIX2DOS,
		},
		Condition: func(cond string) (bool, error) {
			switch cond {
			case "darwin":
				return runtime.GOOS == "darwin", nil
			case "freebsd":
				return runtime.GOOS == "freebsd", nil
			case "githubactionsonwindows":
				return chezmoitest.GitHubActionsOnWindows(), nil
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
// if it does not already exist. It updates the HOME environment variable, and,
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
		chezmoiConfigDir = path.Join(homeDir, ".config", "chezmoi")
		chezmoiSourceDir = path.Join(homeDir, ".local", "share", "chezmoi")
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
	if err != nil || os.FileMode(mode64).Perm() != os.FileMode(mode64) {
		ts.Fatalf("invalid mode: %s", args[0])
	}
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(args[1])
	if err != nil {
		ts.Fatalf("%s: %v", args[1], err)
	}
	equal := info.Mode().Perm() == os.FileMode(mode64)&^chezmoitest.Umask
	if neg && equal {
		ts.Fatalf("%s unexpectedly has mode %03o", args[1], info.Mode().Perm())
	}
	if !neg && !equal {
		ts.Fatalf("%s has mode %03o, expected %03o", args[1], info.Mode().Perm(), os.FileMode(mode64)&^chezmoitest.Umask)
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

// cmdMkAGEConfig creates a AGE key and a chezmoi configuration file.
func cmdMkAGEConfig(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unupported: ! mkageconfig")
	}
	if len(args) > 0 {
		ts.Fatalf("usage: mkageconfig")
	}
	homeDir := ts.Getenv("HOME")
	ts.Check(os.MkdirAll(homeDir, 0o777))
	privateKeyFile := filepath.Join(homeDir, "key.txt")
	publicKey, _, err := chezmoitest.AGEGenerateKey(ts.MkAbs(privateKeyFile))
	ts.Check(err)
	configFile := filepath.Join(homeDir, ".config", "chezmoi", "chezmoi.toml")
	ts.Check(os.MkdirAll(filepath.Dir(configFile), 0o777))
	ts.Check(ioutil.WriteFile(configFile, []byte(fmt.Sprintf(chezmoitest.JoinLines(
		`encryption = "age"`,
		`[age]`,
		`  identity = %q`,
		`  recipient = %q`,
	), privateKeyFile, publicKey)), 0o666))
}

// cmdMkGitConfig makes a .gitconfig file in the home directory.
func cmdMkGitConfig(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkgitconfig")
	}
	if len(args) > 1 {
		ts.Fatalf(("usage: mkgitconfig [path]"))
	}
	path := filepath.Join(ts.Getenv("HOME"), ".gitconfig")
	if len(args) > 0 {
		path = ts.MkAbs(args[0])
	}
	ts.Check(os.MkdirAll(filepath.Dir(path), 0o777))
	ts.Check(ioutil.WriteFile(path, []byte(chezmoitest.JoinLines(
		`[core]`,
		`  autocrlf = false`,
		`[user]`,
		`  name = User`,
		`  email = user@example.com`,
	)), 0o666))
}

// cmdMkGPGConfig creates a GPG key and a chezmoi configuration file.
func cmdMkGPGConfig(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unupported: ! mkgpgconfig")
	}
	if len(args) > 0 {
		ts.Fatalf("usage: mkgpgconfig")
	}

	// Create a new directory for GPG. We can't use a subdirectory of the
	// testscript's working directory because on darwin the absolute path can
	// exceed GPG's limit of sockaddr_un.sun_path (107 characters, see man
	// unix(7)). The limit exists because GPG creates a UNIX domain socket in
	// its home directory and UNIX domain socket paths are limited to
	// sockaddr_un.sun_path characters.
	gpgHomeDir, err := ioutil.TempDir("", "chezmoi-test-gpg-homedir")
	ts.Check(err)
	ts.Defer(func() {
		os.RemoveAll(gpgHomeDir)
	})
	if runtime.GOOS != "windows" {
		ts.Check(os.Chmod(gpgHomeDir, 0o700))
	}

	command, err := chezmoitest.GPGCommand()
	ts.Check(err)

	key, passphrase, err := chezmoitest.GPGGenerateKey(command, gpgHomeDir)
	ts.Check(err)

	configFile := filepath.Join(ts.Getenv("HOME"), ".config", "chezmoi", "chezmoi.toml")
	ts.Check(os.MkdirAll(filepath.Dir(configFile), 0o777))
	ts.Check(ioutil.WriteFile(configFile, []byte(fmt.Sprintf(chezmoitest.JoinLines(
		`encryption = "gpg"`,
		`[gpg]`,
		`  args = [`,
		`    "--homedir", %q,`,
		`    "--no-tty",`,
		`    "--passphrase", %q,`,
		`    "--pinentry-mode", "loopback",`,
		`  ]`,
		`  recipient = %q`,
	), gpgHomeDir, passphrase, key)), 0o666))
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
			".create": "# contents of .create\n",
			".dir": map[string]interface{}{
				"file": "# contents of .dir/file\n",
				"subdir": map[string]interface{}{
					"file": "# contents of .dir/subdir/file\n",
				},
			},
			".empty": "",
			".executable": &vfst.File{
				Perm:     0o777,
				Contents: []byte("# contents of .executable\n"),
			},
			".file": "# contents of .file\n",
			".private": &vfst.File{
				Perm:     0o600,
				Contents: []byte("# contents of .private\n"),
			},
			".symlink":  &vfst.Symlink{Target: ".dir/subdir/file"},
			".template": "key = value\n",
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
			"create_dot_create": "# contents of .create\n",
			"dot_dir": map[string]interface{}{
				"file": "# contents of .dir/file\n",
				"subdir": map[string]interface{}{
					"file": "# contents of .dir/subdir/file\n",
				},
			},
			"dot_remove":                "",
			"empty_dot_empty":           "",
			"executable_dot_executable": "# contents of .executable\n",
			"dot_file":                  "# contents of .file\n",
			"private_dot_private":       "# contents of .private\n",
			"symlink_dot_symlink":       ".dir/subdir/file\n",
			"dot_template.tmpl": chezmoitest.JoinLines(
				`key = {{ "value" }}`,
			),
		},
	})
	if err != nil {
		ts.Fatalf("mksourcedir: %v", err)
	}
}

// cmdRmFinalNewline removes final newlines.
func cmdRmFinalNewline(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! rmfinalnewline")
	}
	if len(args) < 1 {
		ts.Fatalf("usage: rmfinalnewline paths...")
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			ts.Fatalf("%s: %v", filename, err)
		}
		if len(data) == 0 || data[len(data)-1] != '\n' {
			continue
		}
		if err := ioutil.WriteFile(filename, data[:len(data)-1], 0o666); err != nil {
			ts.Fatalf("%s: %v", filename, err)
		}
	}
}

// cmdUNIX2DOS converts files from UNIX line endings to DOS line endings.
func cmdUNIX2DOS(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! unix2dos")
	}
	if len(args) < 1 {
		ts.Fatalf("usage: unix2dos paths...")
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		data, err := ioutil.ReadFile(filename)
		ts.Check(err)
		dosData, err := unix2DOS(data)
		ts.Check(err)
		if err := ioutil.WriteFile(filename, dosData, 0o666); err != nil {
			ts.Fatalf("%s: %v", filename, err)
		}
	}
}

func prependDirToPath(dir, path string) string {
	return strings.Join(append([]string{dir}, filepath.SplitList(path)...), string(os.PathListSeparator))
}

func setup(env *testscript.Env) error {
	var (
		binDir  = filepath.Join(env.WorkDir, "bin")
		homeDir = filepath.Join(env.WorkDir, "home", "user")
	)

	absHomeDir, err := filepath.Abs(homeDir)
	if err != nil {
		return err
	}
	absSlashHomeDir := filepath.ToSlash(absHomeDir)

	var (
		chezmoiConfigDir = path.Join(absSlashHomeDir, ".config", "chezmoi")
		chezmoiSourceDir = path.Join(absSlashHomeDir, ".local", "share", "chezmoi")
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

	root := make(map[string]interface{})
	switch runtime.GOOS {
	case "windows":
		root["/bin"] = map[string]interface{}{
			// editor.cmd is a non-interactive script that appends "# edited\n"
			// to the end of each file and creates an empty .edited file in each
			// directory.
			"editor.cmd": &vfst.File{
				Contents: []byte(chezmoitest.JoinLines(
					`@echo off`,
					`:loop`,
					`IF EXIST %~s1\NUL (`,
					`	copy /y NUL "%~1\.edited" >NUL`,
					`) ELSE (`,
					`	echo # edited >> "%~1"`,
					`)`,
					`shift`,
					`IF NOT "%~1"=="" goto loop`,
				)),
			},
		}
	default:
		root["/bin"] = map[string]interface{}{
			// editor is a non-interactive script that appends "# edited\n" to
			// the end of each file and creates an empty .edited file in each
			// directory.
			"editor": &vfst.File{
				Perm: 0o755,
				Contents: []byte(chezmoitest.JoinLines(
					`#!/bin/sh`,
					``,
					`for name in $*; do`,
					`    if [ -d $name ]; then`,
					`        touch $name/.edited`,
					`    else`,
					`        echo "# edited" >> $name`,
					`    fi`,
					`done`,
				)),
			},
			// shell is a non-interactive script that appends the directory in
			// which it was launched to $WORK/shell.log.
			"shell": &vfst.File{
				Perm: 0o755,
				Contents: []byte(chezmoitest.JoinLines(
					`#!/bin/sh`,
					``,
					`echo $PWD >> '`+filepath.Join(env.WorkDir, "shell.log")+`'`,
				)),
			},
		}
	}

	return vfst.NewBuilder().Build(vfs.NewPathFS(vfs.OSFS, env.WorkDir), root)
}

// unix2DOS returns data with UNIX line endings converted to DOS line endings.
func unix2DOS(data []byte) ([]byte, error) {
	sb := strings.Builder{}
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		if _, err := sb.Write(s.Bytes()); err != nil {
			return nil, err
		}
		if _, err := sb.WriteString("\r\n"); err != nil {
			return nil, err
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}
