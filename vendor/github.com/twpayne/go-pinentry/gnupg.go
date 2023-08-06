package pinentry

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var gnuPGAgentConfPINEntryProgramRx = regexp.MustCompile(`(?m)^\s*pinentry-program\s+(\S+)`)

// WithBinaryNameFromGnuPGAgentConf sets the name of the pinentry binary by
// reading ~/.gnupg/gpg-agent.conf, if it exists.
func WithBinaryNameFromGnuPGAgentConf() (clientOption ClientOption) {
	clientOption = func(*Client) {}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	data, err := os.ReadFile(filepath.Join(userHomeDir, ".gnupg", "gpg-agent.conf"))
	if err != nil {
		return
	}

	match := gnuPGAgentConfPINEntryProgramRx.FindSubmatch(data)
	if match == nil {
		return
	}

	return func(c *Client) {
		c.binaryName = string(match[1])
	}
}

// WithGPGTTY sets the tty.
func WithGPGTTY() ClientOption {
	if runtime.GOOS == "windows" {
		return nil
	}
	gpgTTY, ok := os.LookupEnv("GPG_TTY")
	if !ok {
		return nil
	}
	return WithCommandf("OPTION %s=%s", OptionTTYName, gpgTTY)
}
