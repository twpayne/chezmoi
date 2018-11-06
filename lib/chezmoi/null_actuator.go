package chezmoi

import "os"

// A NullActuator does nothing.
type NullActuator struct{}

func NewNullActuator() *NullActuator {
	return &NullActuator{}
}

func (a *NullActuator) Chmod(string, os.FileMode) error             { return nil }
func (a *NullActuator) Mkdir(string, os.FileMode) error             { return nil }
func (a *NullActuator) RemoveAll(string) error                      { return nil }
func (a *NullActuator) WriteFile(string, []byte, os.FileMode) error { return nil }
