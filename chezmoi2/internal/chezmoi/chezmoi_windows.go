package chezmoi

import (
	"errors"
	"os"
	"unicode/utf16"

	vfs "github.com/twpayne/go-vfs"
	"golang.org/x/sys/windows"
)

// FQDNHostname returns the machine's fully-qualified DNS domain name.
func FQDNHostname(fs vfs.FS) (string, error) {
	n := uint32(windows.MAX_COMPUTERNAME_LENGTH + 1)
	buf := make([]uint16, n)
	err := windows.GetComputerNameEx(windows.ComputerNameDnsFullyQualified, &buf[0], &n)
	if errors.Is(err, windows.ERROR_MORE_DATA) {
		buf = make([]uint16, n)
		err = windows.GetComputerNameEx(windows.ComputerNameDnsFullyQualified, &buf[0], &n)
	}
	if err != nil {
		return "", err
	}
	return string(utf16.Decode(buf[0:n])), nil
}

// GetUmask returns the umask.
func GetUmask() os.FileMode {
	return os.ModePerm
}

// SetUmask sets the umask.
func SetUmask(umask os.FileMode) {}

// isExecutable returns false on Windows.
func isExecutable(info os.FileInfo) bool {
	return false
}

// isPrivate returns false on Windows.
func isPrivate(info os.FileInfo) bool {
	return false
}

func isSlash(c uint8) bool {
	return c == '\\' || c == '/'
}

// umaskPermEqual returns true on Windows.
func umaskPermEqual(perm1 os.FileMode, perm2 os.FileMode, umask os.FileMode) bool {
	return true
}
