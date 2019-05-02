# Contributing

chezmoi development happens on Github. When contributing, please first [open an
issue](https://github.com/twpayne/chezmoi/issues/new) to discuss the change that
you want to make. Bug reports and documentation improvements are particularly
welcome.

All changes are made via pull requests. In your pull request, please make sure
that:

* The commit messages match chezmoi's convention, specifically that they being
  with a capitalized verb in the imperative and give a short description of what
  the commit does. Detailed information or justification can be optionally
  included in the body of the commit message.

* Commits are logically separate, with no merge or "fixup" commits.

* All tests pass. chezmoi's continuous integration tests include strict checks
  using [`github.com/golangci/golangci-lint`](github.com/golangci/golangci-lint)
  and [`mvdan.cc/gofumpt`](mvdan.cc/gofumpt).

* The branch applies cleanly to `master`.

## Release management

Releases are managed with [goreleaser](https://goreleaser.com/).

To create a new release, push a tag, eg:

    git tag -a v0.1.0 -m "First release"
    git push origin v0.1.0

To run a local "snapshot" build without publishing:

    TRAVIS_BUILD_NUMBER=1 goreleaser --snapshot --rm-dist --debug --skip-publish

## Packaging

If you plan to package chezmoi for your distibution, then note:

* Please enable CGO, if possible. chezmoi can be built and run without CGO, but
  the `.chezmoi.group` template variable may not be set on some systems.

* chezmoi includes an `upgrade` command which attempts to self-upgrade. You can
  remove this command completely by building chezmoi with the `noupgrade` build
  tag.