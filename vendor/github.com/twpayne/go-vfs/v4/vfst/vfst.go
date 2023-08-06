// Package vfst provides helper functions for testing code that uses
// github.com/twpayne/go-vfs.
package vfst

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	vfs "github.com/twpayne/go-vfs/v4"
)

var umask fs.FileMode

// A Dir is a directory with a specified permissions and zero or more Entries.
type Dir struct {
	Perm    fs.FileMode
	Entries map[string]interface{}
}

// A File is a file with a specified permissions and contents.
type File struct {
	Perm     fs.FileMode
	Contents []byte
}

// A Symlink is a symbolic link with a specified target.
type Symlink struct {
	Target string
}

// A Test is a test on an vfs.FS.
type Test func(*testing.T, vfs.FS)

// A PathTest is a test on a specified path in an vfs.FS.
type PathTest func(*testing.T, vfs.FS, string)

// A BuilderOption sets an option on a Builder.
type BuilderOption func(*Builder)

// A Builder populates an vfs.FS.
type Builder struct {
	umask   fs.FileMode
	verbose bool
}

// BuilderUmask sets a builder's umask.
func BuilderUmask(umask fs.FileMode) BuilderOption {
	return func(b *Builder) {
		b.umask = umask
	}
}

// BuilderVerbose sets a builder's verbose flag. When true, the builder will
// log all operations with the standard log package.
func BuilderVerbose(verbose bool) BuilderOption {
	return func(b *Builder) {
		b.verbose = verbose
	}
}

// NewBuilder returns a new Builder with the given options set.
func NewBuilder(options ...BuilderOption) *Builder {
	b := &Builder{
		umask:   umask,
		verbose: false,
	}
	for _, option := range options {
		option(b)
	}
	return b
}

// build is a recursive helper for Build.
func (b *Builder) build(fileSystem vfs.FS, path string, i interface{}) error {
	switch i := i.(type) {
	case []interface{}:
		for _, element := range i {
			if err := b.build(fileSystem, path, element); err != nil {
				return err
			}
		}
		return nil
	case *Dir:
		if parentDir := filepath.Dir(path); parentDir != "." {
			if err := b.MkdirAll(fileSystem, parentDir, 0o777); err != nil {
				return err
			}
		}
		if err := b.Mkdir(fileSystem, path, i.Perm); err != nil {
			return err
		}
		entryNames := make([]string, 0, len(i.Entries))
		for entryName := range i.Entries {
			entryNames = append(entryNames, entryName)
		}
		sort.Strings(entryNames)
		for _, entryName := range entryNames {
			if err := b.build(fileSystem, filepath.Join(path, entryName), i.Entries[entryName]); err != nil {
				return err
			}
		}
		return nil
	case map[string]interface{}:
		if err := b.MkdirAll(fileSystem, path, 0o777); err != nil {
			return err
		}
		entryNames := make([]string, 0, len(i))
		for entryName := range i {
			entryNames = append(entryNames, entryName)
		}
		sort.Strings(entryNames)
		for _, entryName := range entryNames {
			if err := b.build(fileSystem, filepath.Join(path, entryName), i[entryName]); err != nil {
				return err
			}
		}
		return nil
	case map[string]string:
		if err := b.MkdirAll(fileSystem, path, 0o777); err != nil {
			return err
		}
		entryNames := make([]string, 0, len(i))
		for entryName := range i {
			entryNames = append(entryNames, entryName)
		}
		sort.Strings(entryNames)
		for _, entryName := range entryNames {
			if err := b.WriteFile(fileSystem, filepath.Join(path, entryName), []byte(i[entryName]), 0o666); err != nil {
				return err
			}
		}
		return nil
	case *File:
		return b.WriteFile(fileSystem, path, i.Contents, i.Perm)
	case string:
		return b.WriteFile(fileSystem, path, []byte(i), 0o666)
	case []byte:
		return b.WriteFile(fileSystem, path, i, 0o666)
	case *Symlink:
		return b.Symlink(fileSystem, i.Target, path)
	case nil:
		return nil
	default:
		return fmt.Errorf("%s: unsupported type %T", path, i)
	}
}

// Build populates fileSystem from root.
func (b *Builder) Build(fileSystem vfs.FS, root interface{}) error {
	return b.build(fileSystem, "/", root)
}

// Mkdir creates directory path with permissions perm. It is idempotent and
// will not fail if path already exists, is a directory, and has permissions
// perm.
func (b *Builder) Mkdir(fileSystem vfs.FS, path string, perm fs.FileMode) error {
	if info, err := fileSystem.Lstat(path); errors.Is(err, fs.ErrNotExist) {
		if b.verbose {
			log.Printf("mkdir -m 0%o %s", perm&^b.umask, path)
		}
		return fileSystem.Mkdir(path, perm&^b.umask)
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("%s: not a directory", path)
	} else if gotPerm, wantPerm := info.Mode()&fs.ModePerm, perm&^b.umask; !PermEqual(gotPerm, wantPerm) {
		return fmt.Errorf("%s has permissions 0%o, want 0%o", path, gotPerm, wantPerm)
	}
	return nil
}

// MkdirAll creates directory path and any missing parent directories with
// permissions perm. It is idempotent and will not file if path already exists
// and is a directory.
func (b *Builder) MkdirAll(fileSystem vfs.FS, path string, perm fs.FileMode) error {
	// Check path.
	info, err := fileSystem.Lstat(path)
	switch {
	case err != nil && errors.Is(err, fs.ErrNotExist):
		// path does not exist, fallthrough to create.
	case err == nil && info.IsDir():
		// path already exists and is a directory.
		return nil
	case err == nil && !info.IsDir():
		// path already exists, but is not a directory.
		return err
	default:
		// Some other error.
		return err
	}

	// Create path.
	if b.verbose {
		log.Printf("mkdir -p -m 0%o %s", perm&^b.umask, path)
	}
	return vfs.MkdirAll(fileSystem, path, perm&^b.umask)
}

// Symlink creates a symbolic link from newname to oldname. It will create any
// missing parent directories with default permissions. It is idempotent and
// will not fail if the symbolic link already exists and points to oldname.
func (b *Builder) Symlink(fileSystem vfs.FS, oldname, newname string) error {
	// Check newname.
	info, err := fileSystem.Lstat(newname)
	switch {
	case err == nil && info.Mode()&fs.ModeType != fs.ModeSymlink:
		// newname exists, but it's not a symlink.
		return fmt.Errorf("%s: not a symbolic link", newname)
	case err == nil:
		// newname exists, and it's a symlink. Check that it is a symlink to
		// oldname.
		gotTarget, err := fileSystem.Readlink(newname)
		if err != nil {
			return err
		}
		if gotTarget != oldname {
			return fmt.Errorf("%s: has target %s, want %s", newname, gotTarget, oldname)
		}
		return nil
	case errors.Is(err, fs.ErrNotExist):
		// newname does not exist, fallthrough to create.
	default:
		// Some other error, return it.
		return err
	}

	// Create newname.
	if err := b.MkdirAll(fileSystem, filepath.Dir(newname), 0o777); err != nil {
		return err
	}
	if b.verbose {
		log.Printf("ln -s %s %s", oldname, newname)
	}
	return fileSystem.Symlink(oldname, newname)
}

// WriteFile writes file path with contents and permissions perm. It will create
// any missing parent directories with default permissions. It is idempotent and
// will not fail if the file already exists, has contents contents, and
// permissions perm.
func (b *Builder) WriteFile(fileSystem vfs.FS, path string, contents []byte, perm fs.FileMode) error {
	if info, err := fileSystem.Lstat(path); errors.Is(err, fs.ErrNotExist) {
		// fallthrough to fileSystem.WriteFile
	} else if err != nil {
		return err
	} else if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: not a regular file", path)
	} else if gotPerm, wantPerm := info.Mode()&fs.ModePerm, perm&^b.umask; !PermEqual(gotPerm, wantPerm) {
		return fmt.Errorf("%s has permissions 0%o, want 0%o", path, gotPerm, wantPerm)
	} else {
		gotContents, err := fileSystem.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Equal(gotContents, contents) {
			return fmt.Errorf("%s: has contents %v, want %v", path, gotContents, contents)
		}
		return nil
	}
	if err := b.MkdirAll(fileSystem, filepath.Dir(path), 0o777); err != nil {
		return err
	}
	if b.verbose {
		log.Printf("install -m 0%o /dev/null %s", perm&^b.umask, path)
	}
	return fileSystem.WriteFile(path, contents, perm&^b.umask)
}

// runTests recursively runs tests on fileSystem.
func runTests(t *testing.T, fileSystem vfs.FS, name string, test interface{}) {
	t.Helper()
	prefix := ""
	if name != "" {
		prefix = name + "_"
	}
	switch test := test.(type) {
	case Test:
		test(t, fileSystem)
	case []Test:
		for i, test := range test {
			t.Run(prefix+strconv.Itoa(i), func(t *testing.T) {
				//nolint:scopelint
				test(t, fileSystem)
			})
		}
	case map[string]Test:
		testNames := make([]string, 0, len(test))
		for testName := range test {
			testNames = append(testNames, testName)
		}
		sort.Strings(testNames)
		for _, testName := range testNames {
			t.Run(prefix+testName, func(t *testing.T) {
				//nolint:scopelint
				test[testName](t, fileSystem)
			})
		}
	case []interface{}:
		for _, u := range test {
			runTests(t, fileSystem, name, u)
		}
	case map[string]interface{}:
		testNames := make([]string, 0, len(test))
		for testName := range test {
			testNames = append(testNames, testName)
		}
		sort.Strings(testNames)
		for _, testName := range testNames {
			runTests(t, fileSystem, prefix+testName, test[testName])
		}
	case nil:
	default:
		t.Fatalf("%s: unsupported type %T", name, test)
	}
}

// RunTests recursively runs tests on fileSystem.
func RunTests(t *testing.T, fileSystem vfs.FS, name string, tests ...interface{}) {
	t.Helper()
	runTests(t, fileSystem, name, tests)
}

// TestContents returns a PathTest that verifies the contents of the file are
// equal to wantContents.
func TestContents(wantContents []byte) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		if gotContents, err := fileSystem.ReadFile(path); err != nil || !bytes.Equal(gotContents, wantContents) {
			t.Errorf("fileSystem.ReadFile(%q) == %v, %v, want %v, <nil>", path, gotContents, err, wantContents)
		}
	}
}

// TestContentsString returns a PathTest that verifies the contetnts of the
// file are equal to wantContentsStr.
func TestContentsString(wantContentsStr string) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		if gotContents, err := fileSystem.ReadFile(path); err != nil || string(gotContents) != wantContentsStr {
			t.Errorf("fileSystem.ReadFile(%q) == %q, %v, want %q, <nil>", path, gotContents, err, wantContentsStr)
		}
	}
}

// testDoesNotExist is a PathTest that verifies that a file or directory does
// not exist.
var testDoesNotExist = func(t *testing.T, fileSystem vfs.FS, path string) {
	t.Helper()
	_, err := fileSystem.Lstat(path)
	if got, want := errors.Is(err, fs.ErrNotExist), true; got != want {
		t.Errorf("_, err := fileSystem.Lstat(%q); errors.Is(err, fs.ErrNotExist) == %v, want %v", path, got, want)
	}
}

// TestDoesNotExist is a PathTest that verifies that a file or directory does
// not exist.
var TestDoesNotExist PathTest = testDoesNotExist

// TestIsDir is a PathTest that verifies that the path is a directory.
var TestIsDir = TestModeType(fs.ModeDir)

// TestModePerm returns a PathTest that verifies that the path's permissions
// are equal to wantPerm.
func TestModePerm(wantPerm fs.FileMode) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		info, err := fileSystem.Lstat(path)
		if err != nil {
			t.Errorf("fileSystem.Lstat(%q) == %+v, %v, want !<nil>, <nil>", path, info, err)
			return
		}
		if gotPerm := info.Mode() & fs.ModePerm; !PermEqual(gotPerm, wantPerm) {
			t.Errorf("fileSystem.Lstat(%q).Mode()&fs.ModePerm == 0%o, want 0%o", path, gotPerm, wantPerm)
		}
	}
}

// TestModeIsRegular is a PathTest that tests that the path is a regular file.
var TestModeIsRegular = TestModeType(0)

// TestModeType returns a PathTest that verifies that the path's mode type is
// equal to wantModeType.
func TestModeType(wantModeType fs.FileMode) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		info, err := fileSystem.Lstat(path)
		if err != nil {
			t.Errorf("fileSystem.Lstat(%q) == %+v, %v, want !<nil>, <nil>", path, info, err)
			return
		}
		if gotModeType := info.Mode() & fs.ModeType; gotModeType != wantModeType {
			t.Errorf("fileSystem.Lstat(%q).Mode()&fs.ModeType == %v, want %v", path, gotModeType, wantModeType)
		}
	}
}

// TestPath returns a Test that runs pathTests on path.
func TestPath(path string, pathTests ...PathTest) Test {
	return func(t *testing.T, fileSystem vfs.FS) {
		t.Helper()
		for i, pathTest := range pathTests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				//nolint:scopelint
				pathTest(t, fileSystem, path)
			})
		}
	}
}

// TestSize returns a PathTest that tests that path's Size() is equal to
// wantSize.
func TestSize(wantSize int64) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		info, err := fileSystem.Lstat(path)
		if err != nil {
			t.Errorf("fileSystem.Lstat(%q) == %+v, %v, want !<nil>, <nil>", path, info, err)
			return
		}
		if gotSize := info.Size(); gotSize != wantSize {
			t.Errorf("fileSystem.Lstat(%q).Size() == %d, want %d", path, gotSize, wantSize)
		}
	}
}

// TestSymlinkTarget returns a PathTest that tests that path's target is wantTarget.
func TestSymlinkTarget(wantTarget string) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		if gotTarget, err := fileSystem.Readlink(path); err != nil || gotTarget != wantTarget {
			t.Errorf("fileSystem.Readlink(%q) == %q, %v, want %q, <nil>", path, gotTarget, err, wantTarget)
			return
		}
	}
}

// TestMinSize returns a PathTest that tests that path's Size() is at least
// wantMinSize.
func TestMinSize(wantMinSize int64) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		info, err := fileSystem.Lstat(path)
		if err != nil {
			t.Errorf("fileSystem.Lstat(%q) == %+v, %v, want !<nil>, <nil>", path, info, err)
			return
		}
		if gotSize := info.Size(); gotSize < wantMinSize {
			t.Errorf("fileSystem.Lstat(%q).Size() == %d, want >=%d", path, gotSize, wantMinSize)
		}
	}
}
