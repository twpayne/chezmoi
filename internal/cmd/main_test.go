package cmd_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/twpayne/go-vfs/v3"
	"github.com/twpayne/go-vfs/v3/vfst"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
	"github.com/twpayne/chezmoi/v2/internal/cmd"
)

var umaskConditionRx = regexp.MustCompile(`\Aumask:([0-7]{3})\z`)

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
			"appendline":     cmdAppendLine,
			"chhome":         cmdChHome,
			"cmpmod":         cmdCmpMod,
			"edit":           cmdEdit,
			"issymlink":      cmdIsSymlink,
			"mkfile":         cmdMkFile,
			"mkageconfig":    cmdMkAGEConfig,
			"mkgitconfig":    cmdMkGitConfig,
			"mkgpgconfig":    cmdMkGPGConfig,
			"mkhomedir":      cmdMkHomeDir,
			"mksourcedir":    cmdMkSourceDir,
			"readlink":       cmdReadLink,
			"removeline":     cmdRemoveLine,
			"rmfinalnewline": cmdRmFinalNewline,
			"unix2dos":       cmdUNIX2DOS,
		},
		Condition: func(cond string) (bool, error) {
			switch cond {
			case "darwin":
				return runtime.GOOS == "darwin", nil
			case "freebsd":
				return runtime.GOOS == "freebsd", nil
			case "linux":
				return runtime.GOOS == "linux", nil
			case "windows":
				return runtime.GOOS == "windows", nil
			}
			if m := umaskConditionRx.FindStringSubmatch(cond); m != nil {
				umask, _ := strconv.ParseInt(m[1], 8, 64)
				return chezmoitest.Umask == fs.FileMode(umask), nil
			}
			return false, fmt.Errorf("%s: unknown condition", cond)
		},
		Setup: setup,
	})
}

// cmdAppendLine appends lines to a file.
func cmdAppendLine(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! appendline")
	}
	if len(args) != 2 {
		ts.Fatalf("usage: appendline file line")
	}
	filename := ts.MkAbs(args[0])
	data, err := os.ReadFile(filename)
	ts.Check(err)
	data = append(data, append([]byte(args[1]), '\n')...)
	ts.Check(os.WriteFile(filename, data, 0o666))
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
	if err != nil || fs.FileMode(mode64).Perm() != fs.FileMode(mode64) {
		ts.Fatalf("invalid mode: %s", args[0])
	}
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(args[1])
	if err != nil {
		ts.Fatalf("%s: %v", args[1], err)
	}
	equal := info.Mode().Perm() == fs.FileMode(mode64)&^chezmoitest.Umask
	if neg && equal {
		ts.Fatalf("%s unexpectedly has mode %03o", args[1], info.Mode().Perm())
	}
	if !neg && !equal {
		ts.Fatalf("%s has mode %03o, expected %03o", args[1], info.Mode().Perm(), fs.FileMode(mode64)&^chezmoitest.Umask)
	}
}

// cmdEdit edits all of its arguments by appending "# edited\n" to them.
func cmdEdit(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! edit")
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		data, err := os.ReadFile(filename)
		if err != nil {
			ts.Fatalf("edit: %v", err)
		}
		data = append(data, []byte("# edited\n")...)
		if err := os.WriteFile(filename, data, 0o666); err != nil {
			ts.Fatalf("edit: %v", err)
		}
	}
}

// cmdIsSymlink returns true if all of its arguments are symlinks.
func cmdIsSymlink(ts *testscript.TestScript, neg bool, args []string) {
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		info, err := os.Lstat(filename)
		if err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
		switch isSymlink := info.Mode().Type() == fs.ModeSymlink; {
		case isSymlink && neg:
			ts.Fatalf("%s is a symlink", arg)
		case !isSymlink && !neg:
			ts.Fatalf("%s is not a symlink", arg)
		}
	}
}

// cmdMkFile creates empty files.
func cmdMkFile(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkfile")
	}
	perm := fs.FileMode(0o666)
	if len(args) >= 1 && strings.HasPrefix(args[0], "-perm=") {
		permStr := strings.TrimPrefix(args[0], "-perm=")
		permUint32, err := strconv.ParseUint(permStr, 8, 32)
		if err != nil {
			ts.Fatalf("%s: bad permissions", permStr)
		}
		perm = fs.FileMode(permUint32)
		args = args[1:]
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		_, err := os.Lstat(filename)
		switch {
		case err == nil:
			ts.Fatalf("%s: already exists", arg)
		case !errors.Is(err, fs.ErrNotExist):
			ts.Fatalf("%s: %v", arg, err)
		}
		if err := os.WriteFile(filename, nil, perm); err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
	}
}

// cmdMkAGEConfig creates a AGE key and a chezmoi configuration file.
func cmdMkAGEConfig(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkageconfig")
	}
	if len(args) > 1 || len(args) == 1 && args[0] != "-symmetric" {
		ts.Fatalf("usage: mkageconfig [-symmetric]")
	}
	symmetric := len(args) == 1 && args[0] == "-symmetric"
	homeDir := ts.Getenv("HOME")
	ts.Check(os.MkdirAll(homeDir, 0o777))
	privateKeyFile := filepath.Join(homeDir, "key.txt")
	publicKey, _, err := chezmoitest.AGEGenerateKey(ts.MkAbs(privateKeyFile))
	ts.Check(err)
	configFile := filepath.Join(homeDir, ".config", "chezmoi", "chezmoi.toml")
	ts.Check(os.MkdirAll(filepath.Dir(configFile), 0o777))
	lines := []string{
		`encryption = "age"`,
		`[age]`,
		`    identity = ` + strconv.Quote(privateKeyFile),
	}
	if symmetric {
		lines = append(lines, `    symmetric = true`)
	} else {
		lines = append(lines, `    recipient = `+strconv.Quote(publicKey))
	}
	ts.Check(os.WriteFile(configFile, []byte(chezmoitest.JoinLines(lines...)), 0o666))
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
	ts.Check(os.WriteFile(path, []byte(chezmoitest.JoinLines(
		`[core]`,
		`    autocrlf = false`,
		`[init]`,
		`    defaultBranch = master`,
		`[user]`,
		`    name = User`,
		`    email = user@example.com`,
	)), 0o666))
}

// cmdMkGPGConfig creates a GPG key and a chezmoi configuration file.
func cmdMkGPGConfig(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkgpgconfig")
	}
	if len(args) > 1 || len(args) == 1 && args[0] != "-symmetric" {
		ts.Fatalf("usage: mkgpgconfig [-symmetric]")
	}
	symmetric := len(args) == 1 && args[0] == "-symmetric"

	// Create a new directory for GPG. We can't use a subdirectory of the
	// testscript's working directory because on darwin the absolute path can
	// exceed GPG's limit of sockaddr_un.sun_path (107 characters, see man
	// unix(7)). The limit exists because GPG creates a UNIX domain socket in
	// its home directory and UNIX domain socket paths are limited to
	// sockaddr_un.sun_path characters.
	gpgHomeDir, err := os.MkdirTemp("", "test-gpg-homedir")
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
	lines := []string{
		`encryption = "gpg"`,
		`[gpg]`,
		`    args = [`,
		`        "--homedir", ` + strconv.Quote(gpgHomeDir) + `,`,
		`        "--no-tty",`,
		`        "--passphrase", ` + strconv.Quote(passphrase) + `,`,
		`        "--pinentry-mode", "loopback",`,
		`    ]`,
	}
	if symmetric {
		lines = append(lines, `    symmetric = true`)
	} else {
		lines = append(lines, `    recipient = "`+key+`"`)
	}
	ts.Check(os.WriteFile(configFile, []byte(chezmoitest.JoinLines(lines...)), 0o666))
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

// cmdReadLink reads a symlink and verifies that its target is as expected.
func cmdReadLink(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: readlink path target")
	}
	filename := ts.MkAbs(args[0])
	link, err := os.Readlink(filename)
	ts.Check(err)
	switch {
	case !neg && link != args[1]:
		ts.Fatalf("readlink: %s -> %s, expected %s", args[0], link, args[1])
	case neg && link == args[1]:
		ts.Fatalf("readlink: %s -> %s, expected ! %s", args[0], link, args[1])
	}
}

// cmdRemoveLine removes lines matching line from file, which must be present.
func cmdRemoveLine(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! removeline")
	}
	if len(args) != 2 {
		ts.Fatalf("usage: removeline file line")
	}
	filename := ts.MkAbs(args[0])
	data, err := os.ReadFile(filename)
	ts.Check(err)
	lineSlice := []byte(args[1])
	lines := bytes.Split(data, []byte{'\n'})
	n := 0
	for _, line := range lines {
		if bytes.Equal(line, lineSlice) {
			continue
		}
		lines[n] = line
		n++
	}
	if n == len(lines) {
		ts.Fatalf("removeline: %q not found in %s", args[1], args[0])
	}
	data = append(bytes.Join(lines[:n], []byte{'\n'}), '\n')
	ts.Check(os.WriteFile(filename, data, 0o666))
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
		data, err := os.ReadFile(filename)
		if err != nil {
			ts.Fatalf("%s: %v", filename, err)
		}
		if len(data) == 0 || data[len(data)-1] != '\n' {
			continue
		}
		if err := os.WriteFile(filename, data[:len(data)-1], 0o666); err != nil {
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
		data, err := os.ReadFile(filename)
		ts.Check(err)
		dosData, err := unix2DOS(data)
		ts.Check(err)
		if err := os.WriteFile(filename, dosData, 0o666); err != nil {
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
