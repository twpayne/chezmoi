package chezmoi

import "os"

// A NullActuator does nothing.
type NullActuator struct{}

// NewNullActuator returns a new NullActuator.
func NewNullActuator() *NullActuator {
	return &NullActuator{}
}

// Chmod implements Actuator.Chmod.
func (a *NullActuator) Chmod(string, os.FileMode) error {
	return nil
}

// Mkdir implements Actuator.Mkdir.
func (a *NullActuator) Mkdir(string, os.FileMode) error {
	return nil
}

// RemoveAll implements Actuator.RemoveAll.
func (a *NullActuator) RemoveAll(string) error {
	return nil
}

// Rename implements Actuator.Rename.
func (a *NullActuator) Rename(string, string) error {
	return nil
}

// WriteFile implements Actuator.WriteFile.
func (a *NullActuator) WriteFile(string, []byte, os.FileMode, []byte) error {
	return nil
}

// WriteSymlink implements Actuator.WriteSymlink.
func (a *NullActuator) WriteSymlink(string, string) error {
	return nil
}
