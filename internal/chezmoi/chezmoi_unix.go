// +build !windows

package chezmoi

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"strings"
	"syscall"

	vfs "github.com/twpayne/go-vfs/v2"
)

var whitespaceRx = regexp.MustCompile(`\s+`)

func init() {
	Umask = os.FileMode(syscall.Umask(0))
	syscall.Umask(int(Umask))
}

// FQDNHostname returns the FQDN hostname from parsing /etc/hosts.
func FQDNHostname(fs vfs.FS) (string, error) {
	etcHostsContents, err := fs.ReadFile("/etc/hosts")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(bytes.NewReader(etcHostsContents))
	for s.Scan() {
		text := s.Text()
		text = strings.TrimSpace(text)
		if index := strings.IndexByte(text, '#'); index != -1 {
			text = text[:index]
		}
		fields := whitespaceRx.Split(text, -1)
		if len(fields) >= 2 && fields[0] == "127.0.1.1" {
			return fields[1], nil
		}
	}
	return "", s.Err()
}

// isExecutable returns if info is executable.
func isExecutable(info os.FileInfo) bool {
	return info.Mode().Perm()&0o111 != 0
}

// isPrivate returns if info is private.
func isPrivate(info os.FileInfo) bool {
	return info.Mode().Perm()&0o77 == 0
}
