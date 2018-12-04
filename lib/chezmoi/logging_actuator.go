package chezmoi

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// A LoggingActuator wraps an Actuator and logs all of the actions it executes
// and any errors.
type LoggingActuator struct {
	w io.Writer
	a Actuator
}

// NewLoggingActuator returns a new LoggingActuator.
func NewLoggingActuator(w io.Writer, a Actuator) *LoggingActuator {
	return &LoggingActuator{
		a: a,
		w: w,
	}
}

// Chmod implements Actuator.Chmod.
func (a *LoggingActuator) Chmod(name string, mode os.FileMode) error {
	action := fmt.Sprintf("chmod %o %s", mode, name)
	err := a.a.Chmod(name, mode)
	if err == nil {
		_, _ = fmt.Fprintln(a.w, action)
	} else {
		_, _ = fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// Mkdir implements Actuator.Mkdir.
func (a *LoggingActuator) Mkdir(name string, perm os.FileMode) error {
	action := fmt.Sprintf("mkdir -m %o %s", perm, name)
	err := a.a.Mkdir(name, perm)
	if err == nil {
		_, _ = fmt.Fprintln(a.w, action)
	} else {
		_, _ = fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// RemoveAll implements Actuator.RemoveAll.
func (a *LoggingActuator) RemoveAll(name string) error {
	action := fmt.Sprintf("rm -rf %s", name)
	err := a.a.RemoveAll(name)
	if err == nil {
		_, _ = fmt.Fprintln(a.w, action)
	} else {
		_, _ = fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// Rename implements Actuator.Rename.
func (a *LoggingActuator) Rename(oldpath, newpath string) error {
	action := fmt.Sprintf("mv %s %s", oldpath, newpath)
	err := a.a.Rename(oldpath, newpath)
	if err == nil {
		_, _ = fmt.Fprintln(a.w, action)
	} else {
		_, _ = fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// WriteFile implements Actuator.WriteFile.
func (a *LoggingActuator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	action := fmt.Sprintf("install -m %o /dev/null %s", perm, name)
	err := a.a.WriteFile(name, data, perm, currData)
	if err == nil {
		_, _ = fmt.Fprintln(a.w, action)
		for _, section := range diff(splitLines(currData), splitLines(data)) {
			for _, s := range section.s {
				_, _ = fmt.Fprintf(a.w, "%c%s\n", section.ctype, s)
			}
		}
	} else {
		_, _ = fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// WriteSymlink implements Actuator.WriteSymlink.
func (a *LoggingActuator) WriteSymlink(oldname, newname string) error {
	action := fmt.Sprintf("ln -sf %s %s", oldname, newname)
	err := a.a.WriteSymlink(oldname, newname)
	if err == nil {
		_, _ = fmt.Fprintln(a.w, action)
	} else {
		_, _ = fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

func splitLines(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	return strings.Split(string(data), "\n")
}
