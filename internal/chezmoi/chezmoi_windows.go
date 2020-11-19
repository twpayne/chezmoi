package chezmoi

import "os"

// GetUmask returns the umask.
func GetUmask() os.FileMode {
	return os.FileMode(0)
}

// SetUmask sets the umask.
func SetUmask(os.FileMode) {}
