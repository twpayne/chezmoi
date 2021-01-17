package chezmoi

import (
	"os"

	vfs "github.com/twpayne/go-vfs"
)

// FQDNHostname does nothing on Windows.
func FQDNHostname(fs vfs.FS) (string, error) {
	// LATER find out how to determine the FQDN hostname on Windows
	return "", nil
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
