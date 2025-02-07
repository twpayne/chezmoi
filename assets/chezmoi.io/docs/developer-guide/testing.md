# Testing

chezmoi uses multiple levels of testing:

1. Unit testing, using [`testing`][testing],  and
   [`github.com/alecthomas/assert/v2`][assert], tests that functions and small
   components behave as expected for a wide range of inputs, especially edge
   cases. These are generally found in `internal/chezmoi/*_test.go`.

2. File system integration tests, using `testing` and
   [`github.com/twpayne/go-vfs/v5`][vfs], test chezmoi's effects on the file
   system. This include some tests in `internal/chezmoi/*_test.go`, and higher
   level command tests in `internal/cmd/*cmd_test.go`.

3. High-level integration tests using
   [`github.com/rogpeppe/go-internal/testscript`][testscript] are in
   `internal/cmd/testdata/scripts/*.txtar` and are run by
   `internal/cmd/main_test.go`.

4. Linux distribution and OS tests run the full test suite using Docker for
   different Linux distributions (in `assets/docker`) and Vagrant for different
   OSes (in `assets/vagrant`). Windows tests are run in GitHub Actions.

[testing]: https://pkg.go.dev/testing
[assert]: https://pkg.go.dev/github.com/alecthomas/assert/v2
[vfs]: https://pkg.go.dev/github.com/twpayne/go-vfs/v5
[testscript]: https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript
