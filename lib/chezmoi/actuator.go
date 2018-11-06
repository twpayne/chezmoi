package chezmoi

import "os"

// An Actuator makes changes.
type Actuator interface {
	Chmod(string, os.FileMode) error
	Mkdir(string, os.FileMode) error
	RemoveAll(string) error
	WriteFile(string, []byte, os.FileMode, []byte) error
}
