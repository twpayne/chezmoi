package chezmoi

import "os"

// GetUmask returns the umask.
func GetUmask() os.FileMode {
	return os.ModePerm
}

// SetUmask sets the umask.
func SetUmask(os.FileMode) {}
