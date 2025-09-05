# Contributing changes

Bug reports, bug fixes, and documentation improvements are always welcome.
Please [open an issue][issue] or [create a pull request][pr] with your report,
fix, or improvement.

If you want to make a more significant change, please first
[open an issue][issue] to discuss the change that you want to make. Dave Cheney
gives a [good rationale][rationale] as to why this is important.

All changes are made via pull requests. In your pull request, please make sure
that:

* All existing tests pass. You can ensure this by running `make test`.

* There are appropriate additional tests that demonstrate that your PR works as
  intended.

* The documentation is updated, if necessary. For new features you should add an
  entry in `assets/chezmoi.io/docs/user-guide/` and a complete description in
  `assets/chezmoi.io/docs/reference/`. See the [website][website] for
  instructions on how to build and view a local version of the documentation. By
  default, chezmoi will panic if a flag is undocumented or a long help is
  missing for a command. You can disable this panic during development by
  setting the environment variable `CHEZMOIDEV` to `ignoreflags=1,ignorehelp=1`.
  Once you have documented the command and its flags, run
  `CHEZMOIDEV=ignoreflags=1,ignorehelp=1 go generate` to generate the embedded
  documentation.

* All generated files are up to date. You can ensure this by running `make
  generate` and including any modified files in your commit.

* The code is correctly formatted. You can ensure this by running `make format`.

* The code passes [`golangci-lint`][golangci-lint]. You can ensure this by
  running `make lint`.

* The commit messages follow the [conventional commits specification][commits].
  chezmoi's release notes are generated directly from the commit messages. The
  following criteria can be used to determine the commit type:

      * Small changes, such as fixing a typo or correcting grammar: `chore`
      * Bug fixes: `fix`
      * Extending an existing feature, or adding a new feature: `feat`
      * Adding to, or updating, the documentation: `docs`
      * Anything not covered by the above: `chore`

    Examples can be found in the [commit history][history].

* Commits are logically separate, with no merge or "fixup" commits.

* The branch applies cleanly to `master`.

[issue]: https://github.com/twpayne/chezmoi/issues/new/choose
[pr]: https://help.github.com/en/articles/creating-a-pull-request
[rationale]: https://dave.cheney.net/2019/02/18/talk-then-code
[golangci-lint]: https://github.com/golangci/golangci-lint
[commits]: https://www.conventionalcommits.org/en/v1.0.0/
[history]: https://github.com/twpayne/chezmoi/commits/master/
[website]: /developer-guide/website.md
