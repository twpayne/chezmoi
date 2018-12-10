package chezmoi

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// A LoggingMutator wraps an Mutator and logs all of the actions it executes
// and any errors.
type LoggingMutator struct {
	m Mutator
	w io.Writer
}

// NewLoggingMutator returns a new LoggingMutator.
func NewLoggingMutator(w io.Writer, m Mutator) *LoggingMutator {
	return &LoggingMutator{
		m: m,
		w: w,
	}
}

// Chmod implements Mutator.Chmod.
func (m *LoggingMutator) Chmod(name string, mode os.FileMode) error {
	action := fmt.Sprintf("chmod %o %s", mode, name)
	err := m.m.Chmod(name, mode)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// Mkdir implements Mutator.Mkdir.
func (m *LoggingMutator) Mkdir(name string, perm os.FileMode) error {
	action := fmt.Sprintf("mkdir -m %o %s", perm, name)
	err := m.m.Mkdir(name, perm)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// RemoveAll implements Mutator.RemoveAll.
func (m *LoggingMutator) RemoveAll(name string) error {
	action := fmt.Sprintf("rm -rf %s", name)
	err := m.m.RemoveAll(name)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// Rename implements Mutator.Rename.
func (m *LoggingMutator) Rename(oldpath, newpath string) error {
	action := fmt.Sprintf("mv %s %s", oldpath, newpath)
	err := m.m.Rename(oldpath, newpath)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// WriteFile implements Mutator.WriteFile.
func (m *LoggingMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	action := fmt.Sprintf("install -m %o /dev/null %s", perm, name)
	err := m.m.WriteFile(name, data, perm, currData)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
		for _, section := range diff(splitLines(currData), splitLines(data)) {
			for _, s := range section.s {
				_, _ = fmt.Fprintf(m.w, "%c%s\n", section.ctype, s)
			}
		}
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// WriteSymlink implements Mutator.WriteSymlink.
func (m *LoggingMutator) WriteSymlink(oldname, newname string) error {
	action := fmt.Sprintf("ln -sf %s %s", oldname, newname)
	err := m.m.WriteSymlink(oldname, newname)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

func splitLines(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	return strings.Split(string(data), "\n")
}
