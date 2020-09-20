package chezmoi

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/diff"
	"github.com/pkg/diff/write"
)

// A VerboseMutator wraps an Mutator and logs all of the actions it executes and
// any errors as pseudo shell commands.
type VerboseMutator struct {
	m               Mutator
	w               io.Writer
	colored         bool
	maxDiffDataSize int
}

// NewVerboseMutator returns a new VerboseMutator.
func NewVerboseMutator(w io.Writer, m Mutator, colored bool, maxDiffDataSize int) *VerboseMutator {
	return &VerboseMutator{
		m:               m,
		w:               w,
		colored:         colored,
		maxDiffDataSize: maxDiffDataSize,
	}
}

// Chmod implements Mutator.Chmod.
func (m *VerboseMutator) Chmod(name string, mode os.FileMode) error {
	action := fmt.Sprintf("chmod %o %s", mode, MaybeShellQuote(name))
	err := m.m.Chmod(name, mode)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// IdempotentCmdOutput implements Mutator.IdempotentCmdOutput.
func (m *VerboseMutator) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	action := cmdString(cmd)
	output, err := m.m.IdempotentCmdOutput(cmd)
	if err != nil {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return output, err
}

// Mkdir implements Mutator.Mkdir.
func (m *VerboseMutator) Mkdir(name string, perm os.FileMode) error {
	action := fmt.Sprintf("mkdir -m %o %s", perm, MaybeShellQuote(name))
	err := m.m.Mkdir(name, perm)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// RemoveAll implements Mutator.RemoveAll.
func (m *VerboseMutator) RemoveAll(name string) error {
	action := fmt.Sprintf("rm -rf %s", MaybeShellQuote(name))
	err := m.m.RemoveAll(name)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// Rename implements Mutator.Rename.
func (m *VerboseMutator) Rename(oldpath, newpath string) error {
	action := fmt.Sprintf("mv %s %s", MaybeShellQuote(oldpath), MaybeShellQuote(newpath))
	err := m.m.Rename(oldpath, newpath)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// RunCmd implements Mutator.RunCmd.
func (m *VerboseMutator) RunCmd(cmd *exec.Cmd) error {
	action := cmdString(cmd)
	err := m.m.RunCmd(cmd)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// Stat implements Mutator.Stat.
func (m *VerboseMutator) Stat(name string) (os.FileInfo, error) {
	return m.m.Stat(name)
}

// WriteFile implements Mutator.WriteFile.
func (m *VerboseMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	action := fmt.Sprintf("install -m %o /dev/null %s", perm, MaybeShellQuote(name))
	err := m.m.WriteFile(name, data, perm, currData)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
		// Don't print diffs if either file is binary.
		if isBinary(currData) || isBinary(data) {
			return nil
		}
		// Don't print diffs if either file is too large.
		if m.maxDiffDataSize != 0 {
			if len(currData) > m.maxDiffDataSize || len(data) > m.maxDiffDataSize {
				return nil
			}
		}
		var opts []write.Option
		if m.colored {
			opts = append(opts, write.TerminalColor())
		}
		if err := diff.Text(filepath.Join("a", name), filepath.Join("b", name), string(currData), string(data), m.w, opts...); err != nil {
			return err
		}
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// WriteSymlink implements Mutator.WriteSymlink.
func (m *VerboseMutator) WriteSymlink(oldname, newname string) error {
	action := fmt.Sprintf("ln -sf %s %s", MaybeShellQuote(oldname), MaybeShellQuote(newname))
	err := m.m.WriteSymlink(oldname, newname)
	if err == nil {
		_, _ = fmt.Fprintln(m.w, action)
	} else {
		_, _ = fmt.Fprintf(m.w, "%s: %v\n", action, err)
	}
	return err
}

// cmdString returns a string representation of cmd.
func cmdString(cmd *exec.Cmd) string {
	s := ShellQuoteArgs(append([]string{cmd.Path}, cmd.Args[1:]...))
	if cmd.Dir == "" {
		return s
	}
	return fmt.Sprintf("( cd %s && %s )", MaybeShellQuote(cmd.Dir), s)
}

func isBinary(data []byte) bool {
	return len(data) != 0 && !strings.HasPrefix(http.DetectContentType(data), "text/")
}
