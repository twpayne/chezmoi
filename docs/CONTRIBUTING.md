# chezmoi Contributing Guide

<!--- toc --->
* [Getting started](#getting-started)
* [Developing locally](#developing-locally)
* [Contributing changes](#contributing-changes)
* [Managing releases](#managing-releases)
* [Packaging](#packaging)

## Getting started

chezmoi is written in [Go](https://golang.org) and development happens on
[GitHub](https://github.com). The rest of this document assumes that you've
checked out chezmoi locally.

## Developing locally

chezmoi requires Go 1.13 or later and Go modules enabled. Enable Go modules by
setting the environment variable `GO111MODULE=on`.

chezmoi is a standard Go project, using standard Go tooling, with a few extra
tools. Install these extra tools with:

    make install-tools

Build chezmoi:

    go build .

Run all tests:

    go test ./...

Run chezmoi:

    go run .

## Contributing changes

Bug reports, bug fixes, and documentation improvements are always welcome.
Please [open an issue](https://github.com/twpayne/chezmoi/issues/new) or [create
a pull request](https://help.github.com/en/articles/creating-a-pull-request)
with your report, fix, or improvement.

If you want to make a more significant change, please first [open an
issue](https://github.com/twpayne/chezmoi/issues/new) to discuss the change that
you want to make. Dave Cheney gives a [good
rationale](https://dave.cheney.net/2019/02/18/talk-then-code) as to why this is
important.

All changes are made via pull requests. In your pull request, please make sure
that:

* All existing tests pass.

* There are appropriate additional tests that demonstrate that your PR works as
  intended.

* The documentation is updated, if necessary. For new features you should add an
  entry in `docs/HOWTO.md` and a complete description in `docs/REFERENCE.md`.

* The code is correctly formatted, according to
  [gofumports](https://mvdan.cc/gofumpt/gofumports). You can ensure this by
  running `make format`.

* The code passes [`go vet`](https://golang.org/cmd/vet/) and
  [`golangci-lint`](https://github.com/golangci/golangci-lint). You can ensure
  this by running `make lint`.

* The commit messages match chezmoi's convention, specifically that they begin
  with a capitalized verb in the imperative and give a short description of what
  the commit does. Detailed information or justification can be optionally
  included in the body of the commit message.

* Commits are logically separate, with no merge or "fixup" commits.

* The branch applies cleanly to `master`.

## Managing releases

Releases are managed with [goreleaser](https://goreleaser.com/).

To create a new release, push a tag, eg:

    git tag -a v0.1.0 -m "First release"
    git push origin v0.1.0

To run a local "snapshot" build without publishing:

    make test-release

## Packaging

If you plan to package chezmoi for your distibution, then note:

* Please set the version number, git commit, and build time in the binary. This is done by passing the linker flags:

  ```
  -X github.com/twpayne/chezmoi/cmd.VersionStr={{ .Version }}
  -X github.com/twpayne/chezmoi/cmd.Commit={{ .Commit }}
  -X github.com/twpayne/chezmoi/cmd.Date={{ .Date }}
  ```

* Please enable CGO, if possible. chezmoi can be built and run without CGO, but
  the `.chezmoi.group` template variable may not be set on some systems.

* chezmoi includes a `docs` command which prints its documentation. By default,
  the docs are embedded in the binary. You can disable this behaviour, and have
  chezmoi read its docs from the filesystem by building with the `noembeddocs`
  build tag and setting the directory where chezmoi can find them with the `-X
  github.com/twpayne/chezmoi/cmd.DocDir={{ .PathToDocs }}` linker flag. For
  example:

  ```
  go build -tags noembeddocs -ldflags "-X github.com/twpayne/chezmoi/cmd.DocsDir=/usr/share/doc/chezmoi" .
  ```

  To disable the `docs` command completely, use the `nodocs` build tag.

* chezmoi includes an `upgrade` command which attempts to self-upgrade. You can
  remove this command completely by building chezmoi with the `noupgrade` build
  tag.

* chezmoi includes shell completions in the `completions` directory. Please
  include these in the package and install them in the shell-appropriate
  directory, if possible.
