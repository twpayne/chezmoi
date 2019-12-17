package chezmoi

import (
	"log"
	"os"
	"os/exec"
	"time"
)

// A DebugMutator wraps a Mutator and logs all of the actions it executes.
type DebugMutator struct {
	m Mutator
}

// NewDebugMutator returns a new DebugMutator.
func NewDebugMutator(m Mutator) *DebugMutator {
	return &DebugMutator{
		m: m,
	}
}

// Chmod implements Mutator.Chmod.
func (m *DebugMutator) Chmod(name string, mode os.FileMode) error {
	return Debugf("Chmod(%q, 0%o)", []interface{}{name, mode}, func() error {
		return m.m.Chmod(name, mode)
	})
}

// IdempotentCmdOutput implements Mutator.IdempotentCmdOutput.
func (m *DebugMutator) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	var output []byte
	cmdStr := ShellQuoteArgs(append([]string{cmd.Path}, cmd.Args[1:]...))
	err := Debugf("IdempotentCmdOutput(%q)", []interface{}{cmdStr}, func() error {
		var err error
		output, err = m.m.IdempotentCmdOutput(cmd)
		return err
	})
	return output, err
}

// Mkdir implements Mutator.Mkdir.
func (m *DebugMutator) Mkdir(name string, perm os.FileMode) error {
	return Debugf("Mkdir(%q, 0%o)", []interface{}{name, perm}, func() error {
		return m.m.Mkdir(name, perm)
	})
}

// RemoveAll implements Mutator.RemoveAll.
func (m *DebugMutator) RemoveAll(name string) error {
	return Debugf("RemoveAll(%q)", []interface{}{name}, func() error {
		return m.m.RemoveAll(name)
	})
}

// Rename implements Mutator.Rename.
func (m *DebugMutator) Rename(oldpath, newpath string) error {
	return Debugf("Rename(%q, %q)", []interface{}{oldpath, newpath}, func() error {
		return m.Rename(oldpath, newpath)
	})
}

// RunCmd implements Mutator.RunCmd.
func (m *DebugMutator) RunCmd(cmd *exec.Cmd) error {
	cmdStr := ShellQuoteArgs(append([]string{cmd.Path}, cmd.Args[1:]...))
	return Debugf("Run(%q)", []interface{}{cmdStr}, func() error {
		return m.m.RunCmd(cmd)
	})
}

// Stat implements Mutator.Stat.
func (m *DebugMutator) Stat(name string) (os.FileInfo, error) {
	var fi os.FileInfo
	err := Debugf("Stat(%q)", []interface{}{name}, func() error {
		var err error
		fi, err = m.m.Stat(name)
		return err
	})
	return fi, err
}

// WriteFile implements Mutator.WriteFile.
func (m *DebugMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	return Debugf("WriteFile(%q, _, 0%o, _)", []interface{}{name, perm}, func() error {
		return m.m.WriteFile(name, data, perm, currData)
	})
}

// WriteSymlink implements Mutator.WriteSymlink.
func (m *DebugMutator) WriteSymlink(oldname, newname string) error {
	return Debugf("WriteSymlink(%q, %q)", []interface{}{oldname, newname}, func() error {
		return m.m.WriteSymlink(oldname, newname)
	})
}

// Debugf logs debugging information about calling f.
func Debugf(format string, args []interface{}, f func() error) error {
	errChan := make(chan error)
	start := time.Now()
	go func(errChan chan<- error) {
		errChan <- f()
	}(errChan)
	select {
	case err := <-errChan:
		if err == nil {
			log.Printf(format+" (%s)", append(args, time.Since(start))...)
		} else {
			log.Printf(format+" == %v (%s)", append(args, err, time.Since(start))...)
		}
		return err
	case <-time.After(1 * time.Second):
		log.Printf(format, args...)
		err := <-errChan
		if err == nil {
			log.Printf(format+" (%s)", append(args, time.Since(start))...)
		} else {
			log.Printf(format+" == %v (%s)", append(args, err, time.Since(start))...)
		}
		return err
	}
}
