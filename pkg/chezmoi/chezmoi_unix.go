//go:build !windows
// +build !windows

package chezmoi

import (
	"bufio"
	"bytes"
	"io/fs"
	"regexp"
	"strings"

	vfs "github.com/twpayne/go-vfs/v4"
	"golang.org/x/sys/unix"
)

var whitespaceRx = regexp.MustCompile(`\s+`)

func init() {
	Umask = fs.FileMode(unix.Umask(0))
	unix.Umask(int(Umask))
}

// FQDNHostname returns the FQDN hostname, if it can be determined.
func FQDNHostname(fileSystem vfs.FS) string {
	if fqdnHostname, err := etcHostsFQDNHostname(fileSystem); err == nil && fqdnHostname != "" {
		return fqdnHostname
	}
	if fqdnHostname, err := etcHostnameFQDNHostname(fileSystem); err == nil && fqdnHostname != "" {
		return fqdnHostname
	}
	return ""
}

// etcHostnameFQDNHostname returns the FQDN hostname from parsing /etc/hostname.
func etcHostnameFQDNHostname(fileSystem vfs.FS) (string, error) {
	contents, err := fileSystem.ReadFile("/etc/hostname")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(bytes.NewReader(contents))
	for s.Scan() {
		text := s.Text()
		text, _, _ = CutString(text, "#")
		if hostname := strings.TrimSpace(text); hostname != "" {
			return hostname, nil
		}
	}
	return "", s.Err()
}

// etcHostsFQDNHostname returns the FQDN hostname from parsing /etc/hosts.
func etcHostsFQDNHostname(fileSystem vfs.FS) (string, error) {
	contents, err := fileSystem.ReadFile("/etc/hosts")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(bytes.NewReader(contents))
	for s.Scan() {
		text := s.Text()
		text = strings.TrimSpace(text)
		text, _, _ = CutString(text, "#")
		fields := whitespaceRx.Split(text, -1)
		if len(fields) >= 2 && fields[0] == "127.0.1.1" {
			return fields[1], nil
		}
	}
	return "", s.Err()
}

// isExecutable returns if fileInfo is executable.
func isExecutable(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o111 != 0
}

// isPrivate returns if fileInfo is private.
func isPrivate(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o77 == 0
}

// isReadOnly returns if fileInfo is read-only.
func isReadOnly(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o222 == 0
}
