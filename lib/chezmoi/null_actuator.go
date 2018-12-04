package chezmoi

import "os"

type nullActuator struct{}

// NullActuator is an Actuator that does nothing.
var NullActuator nullActuator

// Chmod implements Actuator.Chmod.
func (nullActuator) Chmod(string, os.FileMode) error {
	return nil
}

// Mkdir implements Actuator.Mkdir.
func (nullActuator) Mkdir(string, os.FileMode) error {
	return nil
}

// RemoveAll implements Actuator.RemoveAll.
func (nullActuator) RemoveAll(string) error {
	return nil
}

// Rename implements Actuator.Rename.
func (nullActuator) Rename(string, string) error {
	return nil
}

// WriteFile implements Actuator.WriteFile.
func (nullActuator) WriteFile(string, []byte, os.FileMode, []byte) error {
	return nil
}

// WriteSymlink implements Actuator.WriteSymlink.
func (nullActuator) WriteSymlink(string, string) error {
	return nil
}
