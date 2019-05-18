# Contributing

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

chezmoi is a standard Go project, using standard Go tooling.

Build chezmoi:

    go build github.com/twpayne/chezmoi

Run all tests:

    go test github.com/twpayne/chezmoi/...

Run chezmoi:

    go run github.com/twpayne/chezmoi

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

* Your code is correctly formatted, according to
  [gofumports](https://mvdan.cc/gofumpt/gofumports). You can ensure this by
  running `make format`.

* Your code passes [`go vet`](https://golang.org/cmd/vet/) and
  [`golangci-lint`](https://github.com/golangci/golangci-lint). You can ensure
  this by running `make lint`.

* The commit messages match chezmoi's convention, specifically that they being
  with a capitalized verb in the imperative and give a short description of what
  the commit does. Detailed information or justification can be optionally
  included in the body of the commit message.

* Commits are logically separate, with no merge or "fixup" commits.

* All tests pass. chezmoi's continuous integration tests include strict checks
  using [`github.com/golangci/golangci-lint`](github.com/golangci/golangci-lint)
  and [`mvdan.cc/gofumpt`](mvdan.cc/gofumpt).

* The branch applies cleanly to `master`.

## Managing releases

Releases are managed with [goreleaser](https://goreleaser.com/).

Before creating a release, please run:

    make pre-release-checks

This will run a variety of strict checks. Many can be ignored, but please
manually check each of them before tagging a release.

To create a new release, push a tag, eg:

    git tag -a v0.1.0 -m "First release"
    git push origin v0.1.0

To run a local "snapshot" build without publishing:

    make test-release

## Packaging

If you plan to package chezmoi for your distibution, then note:

* Please enable CGO, if possible. chezmoi can be built and run without CGO, but
  the `.chezmoi.group` template variable may not be set on some systems.

* chezmoi includes an `upgrade` command which attempts to self-upgrade. You can
  remove this command completely by building chezmoi with the `noupgrade` build
  tag.