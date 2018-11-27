package chezmoi

import "os"

// An AnyActuator wraps another Actuator and records if any of its methods are
// called.
type AnyActuator struct {
	a        Actuator
	actuated bool
}

// NewAnyActuator returns a new AnyActuator.
func NewAnyActuator(a Actuator) *AnyActuator {
	return &AnyActuator{
		a:        a,
		actuated: false,
	}
}

// Actuated returns true if any of its methods have been called.
func (a *AnyActuator) Actuated() bool {
	return a.actuated
}

// Chmod implements Actuator.Chmod.
func (a *AnyActuator) Chmod(name string, mode os.FileMode) error {
	a.actuated = true
	return a.a.Chmod(name, mode)
}

// Mkdir implements Actuator.Mkdir.
func (a *AnyActuator) Mkdir(name string, perm os.FileMode) error {
	a.actuated = true
	return a.a.Mkdir(name, perm)
}

// RemoveAll implements Actuator.RemoveAll.
func (a *AnyActuator) RemoveAll(name string) error {
	a.actuated = true
	return a.a.RemoveAll(name)
}

// Symlink implements Actuator.Symlink.
func (a *AnyActuator) Symlink(oldname, newname string) error {
	a.actuated = true
	return a.a.Symlink(oldname, newname)
}

// WriteFile implements Actuator.WriteFile.
func (a *AnyActuator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	a.actuated = true
	return a.a.WriteFile(name, data, perm, currData)
}
