# chezmoi contributing guide

<!--- toc --->
* [Getting started](#getting-started)
* [Developing locally](#developing-locally)
* [Generated code](#generated-code)
* [Contributing changes](#contributing-changes)
* [Managing releases](#managing-releases)
* [Packaging](#packaging)
* [Updating the website](#updating-the-website)

## Getting started

chezmoi is written in [Go](https://golang.org) and development happens on
[GitHub](https://github.com). The rest of this document assumes that you've
checked out chezmoi locally.

The [Architecture Guide](ARCHITECTURE.md) contains a high-level overview of
chezmoi's source code.

## Developing locally

chezmoi requires Go 1.16 or later.

chezmoi is a standard Go project, using standard Go tooling, with a few extra
tools. Ensure that these extra tools are installed with:

```console
$ make ensure-tools
```

Build chezmoi:

```console
$ go build .
```

Run all tests:

```console
$ go test ./...
```

Run chezmoi:

```console
$ go run .
```

## Generated code

chezmoi generates shell completions, the install script, and the website from a
single source of truth. You must run

```console
$ go generate
```

if you change includes any of the following:

* Adds or modifies a command.
* Adds or modifies a command's flags.
* Adds or modifies the list of supported OSs and architectures.
* Modifies the install script template.

chezmoi's continuous integration verifies that all generated files are up to
date. Changes to generated files should be included in the commit that modifies
the source of truth.

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

* All generated files are up to date. You can ensure this by running `make
  generate` and including any modified files in your commit.

* The code is correctly formatted, according to
  [`gofumports`](https://mvdan.cc/gofumpt/gofumports). You can ensure this by
  running `make format`.

* The code passes [`golangci-lint`](https://github.com/golangci/golangci-lint).
  You can ensure this by running `make lint`.

* The commit messages match chezmoi's convention, specifically that they begin
  with a capitalized verb in the imperative and give a short description of what
  the commit does. Detailed information or justification can be optionally
  included in the body of the commit message.

* Commits are logically separate, with no merge or "fixup" commits.

* The branch applies cleanly to `master`.

## Managing releases

Releases are managed with [`goreleaser`](https://goreleaser.com/).

To build a test release, without publishing, (Linux only) run:

```console
$ make test-release
```

Publish a new release by creating and pushing a tag, e.g.:

```console
$ git tag v1.2.3
$ git push --tags
```

This triggers a [GitHub Action](https://github.com/twpayne/chezmoi/actions) that
builds and publishes archives, packages, and snaps, and creates a new [GitHub
Release](https://github.com/twpayne/chezmoi/releases).

Publishing [Snaps](https://snapcraft.io/) requires a `SNAPCRAFT_LOGIN`
[repository
secret](https://github.com/twpayne/chezmoi/settings/secrets/actions). Snapcraft
logins periodically expire. Create a new snapcraft login by running:

```console
$ snapcraft export-login --snaps=chezmoi --channels=stable,candidate,beta,edge --acls=package_upload -
```

[brew](https://brew.sh/) formula must be updated manually with the command:

```console
$ brew bump-formula-pr --tag=v1.2.3 chezmoi
```

## Packaging

If you're packaging chezmoi for an operating system or distribution:

* chezmoi has no build or install dependencies other than the standard Go
  tool chain.

* Please set the version number, git commit, and build time in the binary. This
  greatly assists debugging when end users report problems or ask for help. You
  can do this by passing the following flags to the Go linker:

  ```
  -X main.version=$VERSION
  -X main.commit=$COMMIT
  -X main.date=$DATE
  -X main.builtBy=$BUILT_BY
  ```

  `$VERSION` should be the chezmoi version, e.g. `1.7.3`. Any `v` prefix is
  optional and will be stripped, so you can pass the git tag in directly.

  `$COMMIT` should be the full git commit hash at which chezmoi is built, e.g.
  `4d678ce6850c9d81c7ab2fe0d8f20c1547688b91`.

  `$DATE` should be the date of the build in RFC3339 format, e.g.
  `2019-11-23T18:29:25Z`.

  `$BUILT_BY` should be a string indicating what mechanism was used to build the
  binary, e.g. `goreleaser`.

* Please enable cgo, if possible. chezmoi can be built and run without cgo, but
  the `.chezmoi.username` and `.chezmoi.group` template variables may not be set
  correctly on some systems.

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

[The website](https://chezmoi.io) is generated with [Hugo](https://gohugo.io/)
and served with [GitHub pages](https://pages.github.com/) from the [`gh-pages`
branch](https://github.com/twpayne/chezmoi/tree/gh-pages) to GitHub.

Before building the website, you must download the [Hugo Book
Theme](https://github.com/alex-shpak/hugo-book) by running:

```console
$ git submodule update --init
```

Test the website locally by running:

```console
$ ( cd assets/chezmoi.io && hugo serve )
```

and visit http://localhost:1313/.

To build the website in a temporary directory, run:

```console
$ ( cd assets/chezmoi.io && make )
```

From here you can run

```console
$ git show
```

to show changes and

```console
$ git push
```

to push them. You can only push changes if you have write permissions to the
chezmoi GitHub repo.
