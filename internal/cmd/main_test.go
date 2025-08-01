package cmd_test

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/rogpeppe/go-internal/imports"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/chezmoitest"
	"github.com/twpayne/chezmoi/internal/cmd"
)

var (
	envConditionRx   = regexp.MustCompile(`\Aenv:(\w+)\z`)
	envVarRx         = regexp.MustCompile(`\$\w+`)
	lookupRx         = regexp.MustCompile(`\Alookup:(.*)\z`)
	umaskConditionRx = regexp.MustCompile(`\Aumask:([0-7]{3})\z`)

	filterRegex string

	//go:embed mockcommand.tmpl
	mockcommandTmplText string

	//go:embed mockcommand.cmd.tmpl
	mockcommandCmdTmplText string
)

func TestMain(m *testing.M) {
	if strings.Contains(os.Args[0], "cmd.test") {
		flag.StringVar(&filterRegex, "filter", "", "regex to filter test scripts")
		flag.Parse()
	}
	testscript.Main(m, map[string]func(){
		"chezmoi": func() {
			//nolint:revive
			os.Exit(cmd.Main(cmd.VersionInfo{
				Version: "v2.0.0+test",
				Commit:  "HEAD",
				Date:    time.Now().UTC().Format(time.RFC3339),
				BuiltBy: "testscript",
			}, os.Args[1:]))
		},
	})
}

func TestScript(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testscript tests in short mode")
	}
	files, err := filepath.Glob("testdata/scripts/*.txtar")
	if err != nil {
		t.Fatalf("failed to glob files: %v", err)
	}
	if filterRegex != "" {
		re, err := regexp.Compile(filterRegex)
		if err != nil {
			t.Fatalf("invalid regex %q: %v", filterRegex, err)
		}
		var filteredFiles []string
		for _, f := range files {
			baseName := strings.Split(filepath.Base(f), ".")[0]
			if re.MatchString(baseName) {
				filteredFiles = append(filteredFiles, f)
			}
		}
		files = filteredFiles
		if len(files) == 0 {
			t.Fatalf("no test scripts match regex %q", filterRegex)
		}
	}
	testscript.Run(t, testscript.Params{
		Files: files,
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"appendline":     cmdAppendLine,
			"chhome":         cmdChHome,
			"cmpmod":         cmdCmpMod,
			"edit":           cmdEdit,
			"expandenv":      cmdExpandEnv,
			"hexdecode":      cmdHexDecode,
			"httpd":          cmdHTTPD,
			"isdir":          cmdIsDir,
			"issymlink":      cmdIsSymlink,
			"lexists":        cmdLExists,
			"mkfile":         cmdMkFile,
			"mkageconfig":    cmdMkAgeConfig,
			"mkgitconfig":    cmdMkGitConfig,
			"mkgpgconfig":    cmdMkGPGConfig,
			"mkhomedir":      cmdMkHomeDir,
			"mksourcedir":    cmdMkSourceDir,
			"mockcommand":    cmdMockCommand,
			"prependline":    cmdPrependLine,
			"readlink":       cmdReadLink,
			"removeline":     cmdRemoveLine,
			"rmfinalnewline": cmdRmFinalNewline,
			"unix2dos":       cmdUNIX2DOS,
		},
		Condition: func(cond string) (bool, error) {
			if result, valid := goosCondition(cond); valid {
				return result, nil
			}
			if m := envConditionRx.FindStringSubmatch(cond); m != nil {
				return os.Getenv(m[1]) != "", nil
			}
			if m := lookupRx.FindStringSubmatch(cond); m != nil {
				_, err := net.LookupIP(m[1])
				return err == nil, nil
			}
			if m := umaskConditionRx.FindStringSubmatch(cond); m != nil {
				umask, _ := strconv.ParseInt(m[1], 8, 64)
				return chezmoitest.Umask == fs.FileMode(umask), nil
			}
			return false, fmt.Errorf("%s: unknown condition", cond)
		},
		RequireExplicitExec: true,
		RequireUniqueNames:  true,
		Setup:               setup,
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
	ts.Check(os.MkdirAll(homeDir, fs.ModePerm))
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
	fileInfo, err := os.Stat(args[1])
	if err != nil {
		ts.Fatalf("%s: %v", args[1], err)
	}
	equal := fileInfo.Mode().Perm() == fs.FileMode(mode64)&^chezmoitest.Umask
	if neg && equal {
		ts.Fatalf("%s unexpectedly has mode %03o", args[1], fileInfo.Mode().Perm())
	}
	if !neg && !equal {
		format := "%s has mode %03o, expected %03o"
		ts.Fatalf(format, args[1], fileInfo.Mode().Perm(), fs.FileMode(mode64)&^chezmoitest.Umask)
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

// cmdExpandEnv expands environment variables in the given paths.
func cmdExpandEnv(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! expandenv")
	}
	if len(args) == 0 {
		ts.Fatalf("usage: expandenv paths...")
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		data, err := os.ReadFile(filename)
		if err != nil {
			ts.Fatalf("%s: %v", filename, err)
		}
		data = envVarRx.ReplaceAllFunc(data, func(key []byte) []byte {
			if value := ts.Getenv(string(bytes.TrimPrefix(key, []byte{'$'}))); value != "" {
				return []byte(value)
			}
			return key
		})
		if err := os.WriteFile(filename, data, 0o666); err != nil {
			ts.Fatalf("%s: %v", filename, err)
		}
	}
}

// cmdHTTPD starts an HTTP server serving files from the given directory and
// sets the HTTPD_URL environment variable to the URL of the server.
func cmdHTTPD(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! httpd")
	}
	if len(args) != 1 {
		ts.Fatalf("usage: httpd dir")
	}
	dir := ts.MkAbs(args[0])
	server := httptest.NewServer(http.FileServer(http.Dir(dir)))
	ts.Setenv("HTTPD_URL", server.URL)
}

// cmdHexDecode decodes each argument, which must be a file with the extension
// .hex, writing the result to the same path but without the .hex extension.
func cmdHexDecode(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! hexdecode")
	}
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		if filepath.Ext(filename) != ".hex" {
			ts.Fatalf("%s: no .hex extension", arg)
		}
		hexData, err := os.ReadFile(filename)
		if err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
		hexData = regexp.MustCompile(`(?m)#.*$`).ReplaceAll(hexData, nil)
		hexData = regexp.MustCompile(`\s+`).ReplaceAll(hexData, nil)
		data := make([]byte, hex.DecodedLen(len(hexData)))
		if _, err := hex.Decode(data, hexData); err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
		if err := os.WriteFile(strings.TrimSuffix(filename, ".hex"), data, 0o666); err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
	}
}

// cmdIsDir succeeds if all of its arguments are directories.
func cmdIsDir(ts *testscript.TestScript, neg bool, args []string) {
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
		switch isDir := fileInfo.IsDir(); {
		case isDir && neg:
			ts.Fatalf("%s is a directory", arg)
		case !isDir && !neg:
			ts.Fatalf("%s is not a directory", arg)
		}
	}
}

// cmdIsSymlink succeeds if all of its arguments are symlinks.
func cmdIsSymlink(ts *testscript.TestScript, neg bool, args []string) {
	for _, arg := range args {
		filename := ts.MkAbs(arg)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
		switch isSymlink := fileInfo.Mode().Type() == fs.ModeSymlink; {
		case isSymlink && neg:
			ts.Fatalf("%s is a symlink", arg)
		case !isSymlink && !neg:
			ts.Fatalf("%s is not a symlink", arg)
		}
	}
}

// cmdLExists succeeds if all if its arguments exist, without following symlinks.
func cmdLExists(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) == 0 {
		ts.Fatalf("usage: exists file...")
	}

	for _, arg := range args {
		filename := ts.MkAbs(arg)
		switch _, err := os.Lstat(filename); {
		case err == nil && neg:
			ts.Fatalf("%s unexpectedly exists", filename)
		case errors.Is(err, fs.ErrNotExist) && !neg:
			ts.Fatalf("%s does not exist", filename)
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
		switch _, err := os.Lstat(filename); {
		case err == nil:
			ts.Fatalf("%s: already exists", arg)
		case !errors.Is(err, fs.ErrNotExist):
			ts.Fatalf("%s: %v", arg, err)
		}
		if err := writeNewFile(filename, nil, perm); err != nil {
			ts.Fatalf("%s: %v", arg, err)
		}
	}
}

// cmdMkAgeConfig creates an age key and a chezmoi configuration file.
func cmdMkAgeConfig(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkageconfig")
	}
	if len(args) > 1 || len(args) == 1 && args[0] != "-symmetric" {
		ts.Fatalf("usage: mkageconfig [-symmetric]")
	}
	symmetric := len(args) == 1 && args[0] == "-symmetric"
	homeDir := ts.Getenv("HOME")
	ts.Check(os.MkdirAll(homeDir, fs.ModePerm))
	identityFile := filepath.Join(homeDir, "key.txt")
	recipient, err := chezmoitest.AgeGenerateKey("age", ts.MkAbs(identityFile))
	ts.Check(err)
	configFile := filepath.Join(homeDir, ".config", "chezmoi", "chezmoi.toml")
	ts.Check(os.MkdirAll(filepath.Dir(configFile), fs.ModePerm))
	lines := []string{
		`encryption = "age"`,
		`[age]`,
		`    identity = ` + strconv.Quote(identityFile),
	}
	if symmetric {
		lines = append(lines, `    symmetric = true`)
	} else {
		lines = append(lines, `    recipient = `+strconv.Quote(recipient))
	}
	ts.Check(writeNewFile(configFile, []byte(chezmoitest.JoinLines(lines...)), 0o666))
}

// cmdMkGitConfig makes a .gitconfig file in the home directory.
func cmdMkGitConfig(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkgitconfig")
	}
	if len(args) > 1 {
		ts.Fatalf(("usage: mkgitconfig [path]"))
	}
	gitconfigPath := filepath.Join(ts.Getenv("HOME"), ".gitconfig")
	if len(args) > 0 {
		gitconfigPath = ts.MkAbs(args[0])
	}
	ts.Check(os.MkdirAll(filepath.Dir(gitconfigPath), fs.ModePerm))
	ts.Check(writeNewFile(gitconfigPath, []byte(chezmoitest.JoinLines(
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

	command, err := chezmoi.LookPath("gpg")
	ts.Check(err)

	key, passphrase, err := chezmoitest.GPGGenerateKey(command, gpgHomeDir)
	ts.Check(err)

	configFile := filepath.Join(ts.Getenv("HOME"), ".config", "chezmoi", "chezmoi.toml")
	ts.Check(os.MkdirAll(filepath.Dir(configFile), fs.ModePerm))
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
	ts.Check(writeNewFile(configFile, []byte(chezmoitest.JoinLines(lines...)), 0o666))
}

// cmdMkHomeDir makes and populates a home directory.
func cmdMkHomeDir(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mkhomedir")
	}
	if len(args) > 1 {
		ts.Fatalf(("usage: mkhomedir [path]"))
	}
	homeDir := ts.Getenv("HOME")
	if len(args) > 0 {
		homeDir = ts.MkAbs(args[0])
	}
	workDir := ts.Getenv("WORK")
	relPath, err := filepath.Rel(workDir, homeDir)
	ts.Check(err)
	if err := vfst.NewBuilder().Build(vfs.NewPathFS(vfs.OSFS, workDir), map[string]any{
		relPath: map[string]any{
			".create": "# contents of .create\n",
			".dir": map[string]any{
				"file": "# contents of .dir/file\n",
				"subdir": map[string]any{
					"file": "# contents of .dir/subdir/file\n",
				},
			},
			".empty": "",
			".executable": &vfst.File{
				Perm:     fs.ModePerm,
				Contents: []byte("# contents of .executable\n"),
			},
			".file": "# contents of .file\n",
			".private": &vfst.File{
				Perm:     0o600,
				Contents: []byte("# contents of .private\n"),
			},
			".readonly": &vfst.File{
				Perm:     0o444,
				Contents: []byte("# contents of .readonly\n"),
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
	err = vfst.NewBuilder().Build(vfs.NewPathFS(vfs.OSFS, workDir), map[string]any{
		relPath: map[string]any{
			"create_dot_create": "# contents of .create\n",
			"dot_dir": map[string]any{
				"file": "# contents of .dir/file\n",
				"exact_subdir": map[string]any{
					"file": "# contents of .dir/subdir/file\n",
				},
			},
			"dot_remove":                "",
			"empty_dot_empty":           "",
			"executable_dot_executable": "# contents of .executable\n",
			"dot_file":                  "# contents of .file\n",
			"private_dot_private":       "# contents of .private\n",
			"readonly_dot_readonly":     "# contents of .readonly\n",
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

// cmdMockCommand creates a mock command from a definition.
func cmdMockCommand(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! mockcommand")
	}
	if len(args) != 1 {
		ts.Fatalf("usage: mockcommand command")
	}

	command := ts.MkAbs(args[0])
	definitionYAML, err := os.ReadFile(command + ".yaml")
	ts.Check(err)

	// Parse the definition.
	type response struct {
		Args                string            `yaml:"args"`
		OrArgs              []string          `yaml:"orArgs"`
		WindowsArgs         string            `yaml:"windowsArgs"`
		RequireEnv          map[string]string `yaml:"requireEnv"`
		Response            string            `yaml:"response"`
		Destination         string            `yaml:"destination"`
		EscapeChars         bool              `yaml:"escapeChars"`
		SuppressLastNewline bool              `yaml:"suppressLastNewline"`
		ExitCode            int               `yaml:"exitCode"`
	}
	var definition struct {
		Responses []response `yaml:"responses"`
		Default   response   `yaml:"default"`
	}
	ts.Check(chezmoi.FormatYAML.Unmarshal(definitionYAML, &definition))

	// Parse the mock command template.
	var templateName, templateText string
	var renderResponseFunc func(response) string
	switch runtime.GOOS {
	case "windows":
		templateName = "mockcommand.cmd.tmpl"
		templateText = mockcommandCmdTmplText
		escapeCharsRx := regexp.MustCompile(`[\\&|><^]`)
		escapeChars := func(s string) string {
			return escapeCharsRx.ReplaceAllString(s, "^$0")
		}
		renderResponseFunc = func(r response) string {
			var builder strings.Builder
			for _, key := range slices.Sorted(maps.Keys(r.RequireEnv)) {
				value := r.RequireEnv[key]
				fmt.Fprintf(&builder, "    IF NOT \"%%%s%%\" == \"%s\" (\n", key, value)
				fmt.Fprintf(&builder, "        echo.%s=%%%s%%, expected %s\n", key, key, value)
				fmt.Fprintf(&builder, "        exit /b 1\n") //nolint:revive
				fmt.Fprintf(&builder, "    )\n")             //nolint:revive
			}
			var redirect string
			if r.Destination == "stderr" {
				redirect = " 1>&2"
			}
			lines := strings.Split(strings.TrimSuffix(r.Response, "\n"), "\n")
			for i, line := range lines {
				if r.EscapeChars {
					line = escapeChars(line)
				}
				if r.SuppressLastNewline && i == len(lines)-1 {
					fmt.Fprintf(&builder, "    echo | set /p=%s%s\n", line, redirect)
				} else {
					fmt.Fprintf(&builder, "    echo.%s%s\n", line, redirect)
				}
			}
			fmt.Fprintf(&builder, "    exit /b %d", r.ExitCode)
			return builder.String()
		}
	default:
		templateName = "mockcommand.tmpl"
		templateText = mockcommandTmplText
		renderResponseFunc = func(r response) string {
			var builder strings.Builder
			for _, key := range slices.Sorted(maps.Keys(r.RequireEnv)) {
				value := r.RequireEnv[key]
				fmt.Fprintf(&builder, "    if [ \"${%s}\" != \"%s\" ]; then\n", key, value)
				fmt.Fprintf(&builder, "        echo \"%s=${%s}, expected %s\"\n", key, key, value)
				fmt.Fprintf(&builder, "        exit 1\n") //nolint:revive
				fmt.Fprintf(&builder, "    fi\n")         //nolint:revive
			}
			var redirect string
			if r.Destination == "stderr" {
				redirect = " 1>&2"
			}
			if strings.Contains(r.Response, "\n") {
				fmt.Fprintf(&builder, "    cat%s <<EOF\n%sEOF\n", redirect, r.Response)
			} else {
				fmt.Fprintf(&builder, "    echo %q%s\n", r.Response, redirect)
			}
			fmt.Fprintf(&builder, "    exit %d", r.ExitCode)
			return builder.String()
		}
	}
	tmpl, err := template.New(templateName).Funcs(template.FuncMap{
		"default": func(def, s string) string {
			if s == "" {
				return def
			}
			return s
		},
		"quote":          strconv.Quote,
		"renderResponse": renderResponseFunc,
		"replaceAll": func(old, new, s string) string { //nolint:predeclared,revive
			return strings.ReplaceAll(s, old, new)
		},
	}).Parse(templateText)
	ts.Check(err)

	// Create the mock command contents.
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	ts.Check(tmpl.Execute(buffer, definition))
	data := buffer.Bytes()
	if runtime.GOOS == "windows" {
		data = bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
	}

	// Write the mock command.
	switch runtime.GOOS {
	case "windows":
		ts.Check(os.WriteFile(ts.MkAbs(command+".cmd"), data, 0o666))
	default:
		ts.Check(os.WriteFile(ts.MkAbs(command), data, 0o777))
	}
}

// cmdPrependLine prepends lines to a file.
func cmdPrependLine(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! prependline")
	}
	if len(args) != 2 {
		ts.Fatalf("usage: prependline file line")
	}
	filename := ts.MkAbs(args[0])
	data, err := os.ReadFile(filename)
	ts.Check(err)
	data = append(append([]byte(args[1]), '\n'), data...)
	ts.Check(os.WriteFile(filename, data, 0o666))
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

// goosCondition evaluates cond as a logical OR of GOARCHes or GOOSes enclosed
// in parentheses, returning true if any of them match.
func goosCondition(cond string) (result, valid bool) {
	// Interpret the condition as a logical OR of terms in parentheses.
	if !strings.HasPrefix(cond, "(") || !strings.HasSuffix(cond, ")") {
		result = false
		valid = false
		return result, valid
	}
	cond = strings.TrimPrefix(cond, "(")
	cond = strings.TrimSuffix(cond, ")")
	terms := strings.Split(cond, "||")

	// If any of the terms are neither known GOOSes nor GOARCHes then reject the
	// condition as invalid.
	for _, term := range terms {
		if term == "unix" {
			continue
		}
		if _, ok := imports.KnownOS[term]; !ok {
			if _, ok := imports.KnownArch[term]; !ok {
				valid = false
				return result, valid
			}
		}
	}

	// At this point, we know the expression is valid.
	valid = true

	// If any of the terms match either runtime.GOOS or runtime.GOARCH then
	// the condition is true.
	for _, term := range terms {
		switch {
		case term == runtime.GOOS || term == "unix" && imports.UnixOS[runtime.GOOS]:
			result = true
			return result, valid
		case term == runtime.GOARCH:
			result = true
			return result, valid
		}
	}

	// Otherwise, the condition is false.
	result = false
	return result, valid
}

func prependDirToPath(dir, pathStr string) string {
	return strings.Join(append([]string{dir}, filepath.SplitList(pathStr)...), string(os.PathListSeparator))
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

	env.Setenv("HOME", homeDir)
	env.Setenv("PATH", prependDirToPath(binDir, env.Getenv("PATH")))
	if runtime.GOOS == "windows" {
		env.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD;.VBS;.VBE;.JS;.JSE;.WSF;.WSH;.MSC;.CPL")
	}
	env.Setenv("CHEZMOICONFIGDIR", path.Join(absSlashHomeDir, ".config", "chezmoi"))
	env.Setenv("CHEZMOISOURCEDIR", path.Join(absSlashHomeDir, ".local", "share", "chezmoi"))
	env.Setenv("CHEZMOI_GITHUB_TOKEN", os.Getenv("CHEZMOI_GITHUB_TOKEN"))

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

	root := make(map[string]any)
	switch runtime.GOOS {
	case "windows":
		root["/bin"] = map[string]any{
			// editor.cmd is a non-interactive script that appends "# edited\n"
			// to the end of each file and creates an empty .edited file in each
			// directory.
			"editor.cmd": &vfst.File{
				Contents: []byte(chezmoitest.JoinLines(
					`@echo off`,
					`:loop`,
					`IF EXIST %~s1\NUL (`,
					`	copy /y NUL "%~1\.edited" >NUL`,
					// FIXME recursively edit all files if in a directory
					`) ELSE (`,
					`	echo.# edited>>"%~1"`,
					`)`,
					`shift`,
					`IF NOT "%~1"=="" goto loop`,
				)),
			},
		}
	default:
		root["/bin"] = map[string]any{
			// editor is a non-interactive script that appends "# edited\n" to
			// the end of each file.
			"editor": &vfst.File{
				Perm: 0o755,
				Contents: []byte(chezmoitest.JoinLines(
					`#!/bin/sh`,
					``,
					`for name in $*; do`,
					`    if [ -d $name ]; then`,
					`        touch $name/.edited`,
					`        for filename in $(find $name -type f); do`,
					`            echo "# edited" >> $filename`,
					`        done`,
					`    else`,
					`        echo "# edited" >> $name`,
					`    fi`,
					`done`,
				)),
			},
		}
	}

	return vfst.NewBuilder().Build(vfs.NewPathFS(vfs.OSFS, env.WorkDir), root)
}

// unix2DOS returns data with UNIX line endings converted to DOS line endings.
func unix2DOS(data []byte) ([]byte, error) {
	builder := strings.Builder{}
	for line := range bytes.Lines(data) {
		if _, err := builder.Write(line); err != nil {
			return nil, err
		}
		if _, err := builder.WriteString("\r\n"); err != nil {
			return nil, err
		}
	}
	return []byte(builder.String()), nil
}

func writeNewFile(filename string, data []byte, perm fs.FileMode) error {
	switch _, err := os.Lstat(filename); {
	case err == nil:
		return fmt.Errorf("%s: %w", filename, fs.ErrExist)
	case errors.Is(err, fs.ErrNotExist):
		return os.WriteFile(filename, data, perm)
	default:
		return err
	}
}
