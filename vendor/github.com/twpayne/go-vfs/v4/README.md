# go-vfs

[![PkgGoDev](https://pkg.go.dev/badge/github.com/twpayne/go-vfs)](https://pkg.go.dev/github.com/twpayne/go-vfs)
[![Report Card](https://goreportcard.com/badge/github.com/twpayne/go-vfs)](https://goreportcard.com/report/github.com/twpayne/go-vfs)

Package `vfs` provides an abstraction of the `os` and `io` packages that is easy
to test.

## Key features

* File system abstraction layer for commonly-used `os` and `io` functions from
  the standard library.

* Powerful and easy-to-use declarative testing framework, `vfst`. You declare
  the desired state of the filesystem after your code has run, and `vfst` tests
  that the filesystem matches that state. For a quick tour of `vfst`'s features,
  see [the examples in the
  documentation](https://godoc.org/github.com/twpayne/go-vfs/vfst#pkg-examples).

* Compatibility with
  [`github.com/spf13/afero`](https://github.com/spf13/afero) and
  [`github.com/src-d/go-billy`](https://github.com/src-d/go-billy).

## Quick start

`vfs` provides implementations of the `FS` interface:

```go
// An FS is an abstraction over commonly-used functions in the os and io
// packages.
type FS interface {
    Chmod(name string, mode fs.FileMode) error
    Chown(name string, uid, git int) error
    Chtimes(name string, atime, mtime time.Time) error
    Create(name string) (*os.File, error)
    Glob(pattern string) ([]string, error)
    Lchown(name string, uid, git int) error
    Link(oldname, newname string) error
    Lstat(name string) (fs.FileInfo, error)
    Mkdir(name string, perm fs.FileMode) error
    Open(name string) (fs.File, error)
    OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error)
    PathSeparator() rune
    RawPath(name string) (string, error)
    ReadDir(dirname string) ([]fs.DirEntry, error)
    ReadFile(filename string) ([]byte, error)
    Readlink(name string) (string, error)
    Remove(name string) error
    RemoveAll(name string) error
    Rename(oldpath, newpath string) error
    Stat(name string) (fs.FileInfo, error)
    Symlink(oldname, newname string) error
    Truncate(name string, size int64) error
    WriteFile(filename string, data []byte, perm fs.FileMode) error
}
```

To use `vfs`, you write your code to use the `FS` interface, and then use
`vfst` to test it.

`vfs` also provides functions `MkdirAll` (equivalent to `os.MkdirAll`),
`Contains` (an improved `filepath.HasPrefix`), and `Walk` (equivalent to
`filepath.Walk`) that operate on an `FS`.

The implementations of `FS` provided are:

* `OSFS` which calls the underlying `os` and `io` functions directly.

* `PathFS` which transforms all paths to provide a poor-man's `chroot`.

* `ReadOnlyFS` which prevents modification of the underlying FS.

* `TestFS` which assists running tests on a real filesystem but in a temporary
  directory that is easily cleaned up. It uses `OSFS` under the hood.

Example usage:

```go
// writeConfigFile is the function we're going to test. It can make arbitrary
// changes to the filesystem through fileSystem.
func writeConfigFile(fileSystem vfs.FS) error {
    return fileSystem.WriteFile("/home/user/app.conf", []byte(`app config`), 0644)
}

// TestWriteConfigFile is our test function.
func TestWriteConfigFile(t *testing.T) {
    // Create and populate an temporary directory with a home directory.
    fileSystem, cleanup, err := vfst.NewTestFS(map[string]interface{}{
        "/home/user/.bashrc": "# contents of user's .bashrc\n",
    })

    // Check that the directory was populated successfully.
    if err != nil {
        t.Fatalf("vfsTest.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
    }

    // Ensure that the temporary directory is removed.
    defer cleanup()

    // Call the function we want to test.
    if err := writeConfigFile(fileSystem); err != nil {
        t.Error(err)
    }

    // Check properties of the filesystem after our function has modified it.
    vfst.RunTests(t, fileSystem, "app_conf",
        vfst.TestPath("/home/user/app.conf",
            vfst.TestModeIsRegular,
            vfst.TestModePerm(0644),
            vfst.TestContentsString("app config"),
        ),
    )
}
```

## `github.com/spf13/afero` compatibility

There is a compatibility shim for
[`github.com/spf13/afero`](https://github.com/spf13/afero) in
[`github.com/twpayne/go-vfsafero`](https://github.com/twpayne/go-vfsafero). This
allows you to use `vfst` to test existing code that uses
[`afero.FS`](https://godoc.org/github.com/spf13/afero#Fs). See [the
documentation](https://godoc.org/github.com/twpayne/go-vfsafero) for an example.

## `github.com/src-d/go-billy` compatibility

There is a compatibility shim for
[`github.com/src-d/go-billy`](https://github.com/src-d/go-billy) in
[`github.com/twpayne/go-vfsbilly`](https://github.com/twpayne/go-vfsbilly). This
allows you to use `vfst` to test existing code that uses
[`billy.Filesystem`](https://godoc.org/github.com/src-d/go-billy#Filesystem).
See [the documentation](https://godoc.org/github.com/twpayne/go-vfsbilly) for an
example.

## Motivation

`vfs` was inspired by
[`github.com/spf13/afero`](https://github.com/spf13/afero). So, why not use
`afero`?

* `afero` has several critical bugs in its in-memory mock filesystem
  implementation `MemMapFs`, to the point that it is unusable for non-trivial
  test cases. `vfs` does not attempt to implement an in-memory mock filesystem,
  and instead only provides a thin layer around the standard library's `os` and
  `io` packages, and as such should have fewer bugs.

* `afero` does not support creating or reading symbolic links, and its
  `LstatIfPossible` interface is clumsy to use as it is not part of the
  `afero.Fs` interface. `vfs` provides out-of-the-box support for symbolic links
  with all methods in the `FS` interface.

* `afero` has been effectively abandoned by its author, and a "friendly fork"
  ([`github.com/absfs/afero`](https://github.com/absfs/afero)) has not seen much
  activity. `vfs`, by providing much less functionality than `afero`, should be
  smaller and easier to maintain.

## License

MIT
