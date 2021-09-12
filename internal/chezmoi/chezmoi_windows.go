package chezmoi

import (
	"errors"
	"io/fs"
	"unicode/utf16"

	vfs "github.com/twpayne/go-vfs/v4"
	"golang.org/x/sys/windows"
)

// FQDNHostname returns the machine's fully-qualified DNS domain name, if
// available.
func FQDNHostname(fileSystem vfs.FS) string {
	n := uint32(windows.MAX_COMPUTERNAME_LENGTH + 1)
	buf := make([]uint16, n)
	err := windows.GetComputerNameEx(windows.ComputerNameDnsFullyQualified, &buf[0], &n)
	if errors.Is(err, windows.ERROR_MORE_DATA) {
		buf = make([]uint16, n)
		err = windows.GetComputerNameEx(windows.ComputerNameDnsFullyQualified, &buf[0], &n)
	}
	if err != nil {
		return ""
	}
	return string(utf16.Decode(buf[0:n]))
}

// isExecutable returns false on Windows.
func isExecutable(info fs.FileInfo) bool {
	return false
}

// isPrivate returns false on Windows.
func isPrivate(info fs.FileInfo) bool {
	return false
}

// isReadOnly returns false on Windows.
func isReadOnly(info fs.FileInfo) bool {
	return false
}

func isSlash(c uint8) bool {
	return c == '\\' || c == '/'
}
