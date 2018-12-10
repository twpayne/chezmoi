package chezmoi

import "os"

type nullMutator struct{}

// NullMutator is an Mutator that does nothing.
var NullMutator nullMutator

// Chmod implements Mutator.Chmod.
func (nullMutator) Chmod(string, os.FileMode) error {
	return nil
}

// Mkdir implements Mutator.Mkdir.
func (nullMutator) Mkdir(string, os.FileMode) error {
	return nil
}

// RemoveAll implements Mutator.RemoveAll.
func (nullMutator) RemoveAll(string) error {
	return nil
}

// Rename implements Mutator.Rename.
func (nullMutator) Rename(string, string) error {
	return nil
}

// WriteFile implements Mutator.WriteFile.
func (nullMutator) WriteFile(string, []byte, os.FileMode, []byte) error {
	return nil
}

// WriteSymlink implements Mutator.WriteSymlink.
func (nullMutator) WriteSymlink(string, string) error {
	return nil
}
