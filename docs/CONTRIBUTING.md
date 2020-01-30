# chezmoi Contributing Guide

<!--- toc --->
* [Getting started](#getting-started)
* [Developing locally](#developing-locally)
* [Documentation and templates changes](#documentation-and-templates-changes)
* [Contributing changes](#contributing-changes)
* [Managing releases](#managing-releases)
* [Packaging](#packaging)
* [Updating the website](#updating-the-website)

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

## Documentation and templates changes

The canonical documentation for chezmoi is in the `docs` directory. The help
text (the output of `chezmoi command --help`) and the website
(https://chezmoi.io/) are generated from this.

chezmoi embeds documentation and templates in its binary.

If you update any file in the `docs/` or `templates/` directories, you must also run

    go generate

## Contributing changes

Bug reports, bug fixes, and documentation improvements are always welcome.
Please [open an issue](https://github.com/twpayne/chezmoi/issues/new/choose) or
[create a pull
request](https://help.github.com/en/articles/creating-a-pull-request) with your
report, fix, or improvement.

If you want to make a more significant change, please first [open an
issue](https://github.com/twpayne/chezmoi/issues/new/choose) to discuss the
change that you want to make. Dave Cheney gives a [good
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

If you're packaging chezmoi for an operating system or distribution:

* Please set the version number, git commit, and build time in the binary. This
  greatly assists debugging when end users report problems or ask for help. You
  can do this by passing the following flags to the Go linker:

  ```
  -X github.com/twpayne/chezmoi/cmd.VersionStr=$VERSION
  -X github.com/twpayne/chezmoi/cmd.Commit=$COMMIT
  -X github.com/twpayne/chezmoi/cmd.Date=$DATE
  ```

  `$VERSION` should be the chezmoi version, e.g. `1.7.3`. Any `v` prefix is
  optional and will be stripped, so you can pass the git tag in directly.

  `$COMMIT` should be the full git commit hash at which chezmoi is built, e.g.
  `4d678ce6850c9d81c7ab2fe0d8f20c1547688b91`.

  `$DATE` should be the date of the build in RFC3339 format, e.g.
  `2019-11-23T18:29:25Z`.

* Please enable CGO, if possible. chezmoi can be built and run without CGO, but
  the `.chezmoi.username` and `.chezmoi.group` template variables may not be set
  correctly on some systems.

* chezmoi includes a `docs` command which prints its documentation. By default,
  the docs are embedded in the binary. You can disable this behaviour, and have
  chezmoi read its docs from the filesystem by building with the `noembeddocs`
  build tag and setting the directory where chezmoi can find them with the `-X
  github.com/twpayne/chezmoi/cmd.DocDir=$DOCDIR` linker flag. For example:

  ```
  go build -tags noembeddocs -ldflags "-X github.com/twpayne/chezmoi/cmd.DocsDir=/usr/share/doc/chezmoi" .
  ```

  To remove the `docs` command completely, use the `nodocs` build tag.

* chezmoi includes an `upgrade` command which attempts to self-upgrade. You can
  remove this command completely by building chezmoi with the `noupgrade` build
  tag.

* chezmoi includes shell completions in the `completions` directory. Please
  include these in the package and install them in the shell-appropriate
  directory, if possible.

* If the instructions for installing chezmoi in chezmoi's [install
  guide](https://github.com/twpayne/chezmoi/blob/master/docs/INSTALL.md) are
  absent or incorrect, please open an issue or submit a PR to correct them.

## Updating the website

[The website](https://chezmoi.io), is generated with [Hugo](https://gohugo.io/)
and served with [GitHub pages](https://pages.github.com/) from the [`gh-pages`
branch](https://github.com/twpayne/chezmoi/tree/gh-pages) to GitHub.

Before building the website, you must download the [Hugo Book
Theme](https://github.com/alex-shpak/hugo-book) by running:

    git submodule update --init

Test the website locally by running:

    ( cd chezmoi.io && hugo serve )

and visit http://localhost:1313/.

To build the website in a temporary directory, run:

    ( cd chezmoi.io && make )

From here you can run

    git show

to show changes and

    git push

to push them. You can only push changes if you have write permissions to the
chezmoi GitHub repo.