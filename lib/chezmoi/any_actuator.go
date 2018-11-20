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
func (a *AnyActuator) Mkdir(name string, mode os.FileMode) error {
	a.actuated = true
	return a.a.Mkdir(name, mode)
}

// RemoveAll implements Actuator.RemoveAll.
func (a *AnyActuator) RemoveAll(name string) error {
	a.actuated = true
	return a.a.RemoveAll(name)
}

// WriteFile implements Actuator.WriteFile.
func (a *AnyActuator) WriteFile(name string, contents []byte, mode os.FileMode, currentContents []byte) error {
	a.actuated = true
	return a.a.WriteFile(name, contents, mode, currentContents)
}
