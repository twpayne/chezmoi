# Contributing changes

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

* All existing tests pass. You can ensure this by running `make test`.

* There are appropriate additional tests that demonstrate that your PR works as
  intended.

* The documentation is updated, if necessary. For new features you should add an
  entry in `assets/chezmoi.io/docs/user-guide/` and a complete description in
  `assets/chezmoi.io/docs/reference/`. See [website](website.md) for
  instructions on how to build and view a local version of the documentation.

* All generated files are up to date. You can ensure this by running `make
  generate` and including any modified files in your commit.

* The code is correctly formatted, according to
  [`gofumpt`](https://mvdan.cc/gofumpt/). You can ensure this by running `make
  format`.

* The code passes [`golangci-lint`](https://github.com/golangci/golangci-lint).
  You can ensure this by running `make lint`.

* The commit messages follow the [conventional commits
  specification](https://www.conventionalcommits.org/en/v1.0.0/). chezmoi's
  release notes are generated directly from the commit messages. For trivial or
  user-invisible changes, please use the prefix `chore:`.

* Commits are logically separate, with no merge or "fixup" commits.

* The branch applies cleanly to `master`.
