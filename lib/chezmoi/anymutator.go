package chezmoi

import "os"

// An AnyMutator wraps another Mutator and records if any of its mutating
// methods are called.
type AnyMutator struct {
	m       Mutator
	mutated bool
}

// NewAnyMutator returns a new AnyMutator.
func NewAnyMutator(m Mutator) *AnyMutator {
	return &AnyMutator{
		m:       m,
		mutated: false,
	}
}

// IsPrivate implements Mutator.IsPrivate.
func (m *AnyMutator) IsPrivate(file string, umask os.FileMode) bool {
    return m.m.IsPrivate(file, umask)
}

// Chmod implements Mutator.Chmod.
func (m *AnyMutator) Chmod(name string, mode os.FileMode) error {
	m.mutated = true
	return m.m.Chmod(name, mode)
}

// Mkdir implements Mutator.Mkdir.
func (m *AnyMutator) Mkdir(name string, perm os.FileMode) error {
	m.mutated = true
	return m.m.Mkdir(name, perm)
}

// Mutated returns true if any of its methods have been called.
func (m *AnyMutator) Mutated() bool {
	return m.mutated
}

// RemoveAll implements Mutator.RemoveAll.
func (m *AnyMutator) RemoveAll(name string) error {
	m.mutated = true
	return m.m.RemoveAll(name)
}

// Rename implements Mutator.Rename.
func (m *AnyMutator) Rename(oldpath, newpath string) error {
	m.mutated = true
	return m.m.Rename(oldpath, newpath)
}

// Stat implements Mutator.Stat.
func (m *AnyMutator) Stat(path string) (os.FileInfo, error) {
	return m.m.Stat(path)
}

// WriteFile implements Mutator.WriteFile.
func (m *AnyMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	m.mutated = true
	return m.m.WriteFile(name, data, perm, currData)
}

// WriteSymlink implements Mutator.WriteSymlink.
func (m *AnyMutator) WriteSymlink(oldname, newname string) error {
	m.mutated = true
	return m.m.WriteSymlink(oldname, newname)
}
