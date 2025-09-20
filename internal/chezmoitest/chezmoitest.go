// Package chezmoitest contains test helper functions for chezmoi.
package chezmoitest

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"chezmoi.io/chezmoi/internal/chezmoilog"
)

var ageRecipientRx = regexp.MustCompile(`(?m)^Public key: ([0-9a-z]+)\s*$`)

// AgeGenerateKey generates an identity in identityFile and returns the
// recipient.
func AgeGenerateKey(command, identityFile string) (string, error) {
	cmd := exec.Command(command+"-keygen", "--output", identityFile)
	output, err := chezmoilog.LogCmdCombinedOutput(slog.Default(), cmd)
	if err != nil {
		return "", err
	}
	match := ageRecipientRx.FindSubmatch(output)
	if match == nil {
		return "", fmt.Errorf("recipient not found in %q", output)
	}
	return string(match[1]), nil
}

// GPGGenerateKey generates GPG key in homeDir and returns the key and the
// passphrase.
func GPGGenerateKey(command, homeDir string) (key, passphrase string, err error) {
	key = "chezmoi-test-gpg-key"
	passphrase = "chezmoi-test-gpg-passphrase"
	cmd := exec.Command(
		command,
		"--batch",
		"--homedir", homeDir,
		"--no-tty",
		"--passphrase", passphrase,
		"--pinentry-mode", "loopback",
		"--quick-generate-key", key,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = chezmoilog.LogCmdRun(slog.Default(), cmd)
	return key, passphrase, err
}

// HomeDir returns the home directory.
func HomeDir() string {
	switch runtime.GOOS {
	case "windows":
		return "C:/home/user"
	default:
		return "/home/user"
	}
}

// JoinLines joins lines with newlines.
func JoinLines(lines ...string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// SkipUnlessGOOS calls t.Skip() if name does not match runtime.GOOS.
func SkipUnlessGOOS(t *testing.T, name string) {
	t.Helper()
	switch {
	case strings.HasSuffix(name, "_windows") && runtime.GOOS != "windows":
		t.Skip("skipping Windows test on UNIX")
	case strings.HasSuffix(name, "_unix") && runtime.GOOS == "windows":
		t.Skip("skipping UNIX test on Windows")
	}
}

// WithTestFS calls f with a test filesystem populated with root.
func WithTestFS(t *testing.T, root any, f func(vfs.FS)) {
	t.Helper()
	fileSystem, cleanup, err := vfst.NewTestFS(root, vfst.BuilderUmask(Umask))
	assert.NoError(t, err)
	t.Cleanup(cleanup)
	f(fileSystem)
}

// mustParseFileMode parses s as a fs.FileMode and panics on any error.
func mustParseFileMode(s string) fs.FileMode {
	u, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		panic(err)
	}
	return fs.FileMode(uint32(u))
}
