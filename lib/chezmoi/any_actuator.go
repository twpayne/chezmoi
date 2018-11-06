package chezmoi

import "os"

// An AnyActuator records if any of its methods are called.
type AnyActuator struct {
	a        Actuator
	actuated bool
}

func NewAnyActuator(a Actuator) *AnyActuator {
	return &AnyActuator{
		a:        a,
		actuated: false,
	}
}

func (a *AnyActuator) Actuated() bool {
	return a.actuated
}

func (a *AnyActuator) Chmod(name string, mode os.FileMode) error {
	a.actuated = true
	return a.a.Chmod(name, mode)
}

func (a *AnyActuator) Mkdir(name string, mode os.FileMode) error {
	a.actuated = true
	return a.a.Mkdir(name, mode)
}

func (a *AnyActuator) RemoveAll(name string) error {
	a.actuated = true
	return a.a.RemoveAll(name)
}

func (a *AnyActuator) WriteFile(name string, contents []byte, mode os.FileMode) error {
	a.actuated = true
	return a.a.WriteFile(name, contents, mode)
}
