package chezmoi

import "os"

// NullMutator is an Mutator that does nothing.
type NullMutator struct{}

// Chmod implements Mutator.Chmod.
func (NullMutator) Chmod(string, os.FileMode) error {
	return nil
}

// Mkdir implements Mutator.Mkdir.
func (NullMutator) Mkdir(string, os.FileMode) error {
	return nil
}

// RemoveAll implements Mutator.RemoveAll.
func (NullMutator) RemoveAll(string) error {
	return nil
}

// Rename implements Mutator.Rename.
func (NullMutator) Rename(string, string) error {
	return nil
}

// Stat implements Mutator.Stat.
func (NullMutator) Stat(path string) (os.FileInfo, error) {
	return nil, &os.PathError{
		Op:   "stat",
		Path: path,
		Err:  os.ErrNotExist,
	}
}

// WriteFile implements Mutator.WriteFile.
func (NullMutator) WriteFile(string, []byte, os.FileMode, []byte) error {
	return nil
}

// WriteSymlink implements Mutator.WriteSymlink.
func (NullMutator) WriteSymlink(string, string) error {
	return nil
}
