package chezmoi

import (
	"errors"
	"os"
	"unicode/utf16"

	vfs "github.com/twpayne/go-vfs/v2"
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
