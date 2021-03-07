// Package chezmoitest contains test helper functions for chezmoi.
package chezmoitest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v2"
	"github.com/twpayne/go-vfs/v2/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

var (
	agePublicKeyRx                    = regexp.MustCompile(`(?m)^Public key: ([0-9a-z]+)\s*$`)
	gpgKeyMarkedAsUltimatelyTrustedRx = regexp.MustCompile(`(?m)^gpg: key ([0-9A-F]+) marked as ultimately trusted\s*$`)
)

// AGEGenerateKey generates and returns an age public key and the path to the
// private key. If filename is non-zero then the private key is written to it,
// otherwise a new file is created in a temporary directory and the caller is
// responsible for removing the temporary directory.
func AGEGenerateKey(filename string) (publicKey, privateKeyFile string, err error) {
	if filename == "" {
		var tempDir string
		tempDir, err = os.MkdirTemp("", "chezmoi-test-age-key")
		if err != nil {
			return "", "", err
		}
		defer func() {
			if err != nil {
				os.RemoveAll(tempDir)
			}
		}()
		if runtime.GOOS != "windows" {
			if err = os.Chmod(tempDir, 0o700); err != nil {
				return
			}
		}
		filename = filepath.Join(tempDir, "key.txt")
	}

	privateKeyFile = filename
	var output []byte
	cmd := exec.Command("age-keygen", "--output", privateKeyFile)
	output, err = chezmoilog.LogCmdCombinedOutput(log.Logger, cmd)
	if err != nil {
		return
	}
	match := agePublicKeyRx.FindSubmatch(output)
	if match == nil {
		err = fmt.Errorf("public key not found in %q", output)
		return
	}
	publicKey = string(match[1])
	return
}

// GPGCommand returns the GPG command, if it can be found.
func GPGCommand() (string, error) {
	return exec.LookPath("gpg")
}

// GPGGenerateKey generates and returns a GPG key in homeDir.
func GPGGenerateKey(command, homeDir string) (key, passphrase string, err error) {
	//nolint:gosec
	passphrase = "chezmoi-test-gpg-passphrase"
	cmd := exec.Command(
		command,
		"--batch",
		"--homedir", homeDir,
		"--no-tty",
		"--passphrase", passphrase,
		"--pinentry-mode", "loopback",
		"--quick-generate-key", "chezmoi-test-gpg-key",
	)
	output, err := chezmoilog.LogCmdCombinedOutput(log.Logger, cmd)
	if err != nil {
		return "", "", err
	}
	submatch := gpgKeyMarkedAsUltimatelyTrustedRx.FindSubmatch(output)
	if submatch == nil {
		return "", "", fmt.Errorf("key not found in %q", output)
	}
	return string(submatch[1]), passphrase, nil
}

// GitHubActionsOnWindows returns if running in GitHub Actions on Windows.
func GitHubActionsOnWindows() bool {
	return runtime.GOOS == "windows" && os.Getenv("GITHUB_ACTIONS") == "true"
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
func WithTestFS(t *testing.T, root interface{}, f func(fs vfs.FS)) {
	t.Helper()
	fs, cleanup, err := vfst.NewTestFS(root)
	require.NoError(t, err)
	t.Cleanup(cleanup)
	f(fs)
}
