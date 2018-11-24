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
		fmt.Fprintln(a.w, action)
	} else {
		fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// Mkdir implements Actuator.Mkdir.
func (a *LoggingActuator) Mkdir(name string, mode os.FileMode) error {
	action := fmt.Sprintf("mkdir -m %o %s", mode, name)
	err := a.a.Mkdir(name, mode)
	if err == nil {
		fmt.Fprintln(a.w, action)
	} else {
		fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// RemoveAll implements Actuator.RemoveAll.
func (a *LoggingActuator) RemoveAll(name string) error {
	action := fmt.Sprintf("rm -rf %s", name)
	err := a.a.RemoveAll(name)
	if err == nil {
		fmt.Fprintln(a.w, action)
	} else {
		fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

// WriteFile implements Actuator.WriteFile.
func (a *LoggingActuator) WriteFile(name string, contents []byte, mode os.FileMode, currentContents []byte) error {
	action := fmt.Sprintf("install -m %o /dev/null %s", mode, name)
	err := a.a.WriteFile(name, contents, mode, currentContents)
	if err == nil {
		fmt.Fprintln(a.w, action)
		for _, section := range diff(splitLines(currentContents), splitLines(contents)) {
			for _, s := range section.s {
				fmt.Fprintf(a.w, "%c%s\n", section.ctype, s)
			}
		}
	} else {
		fmt.Fprintf(a.w, "%s: %v\n", action, err)
	}
	return err
}

func splitLines(contents []byte) []string {
	if len(contents) == 0 {
		return nil
	}
	return strings.Split(string(contents), "\n")
}
