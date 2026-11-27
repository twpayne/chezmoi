# Contributing changes

!!! warning

    chezmoi has a strict zero-tolerance policy on LLM contributions. If you use
    an LLM (Large Language Model, like ChatGPT, Claude, Gemini, GitHub Copilot,
    or Llama) to make any kind of contribution then you will immediately be
    banned without recourse. See [`CODE_OF_CONDUCT.md`][coc] for more
    information.

Bug reports, bug fixes, and documentation improvements are always welcome.
Please [open an issue][issue] or [create a pull request][pr] with your report,
fix, or improvement.

If you want to make a more significant change, please first
[open an issue][issue] to discuss the change that you want to make. Dave Cheney
gives a [good rationale][rationale] as to why this is important.

## Commit messages

chezmoi's release notes are [automatically generated][release-history-generator]
from its commit messages. For this reason, chezmoi requires strictly-formatted
commit messages, similar to, but not identical to, the [conventional commits
specification][commits].

Four commit message prefixes are permitted:

  * `feat:` for extending an existing user-visible feature or adding a new feature
  * `fix:` for user-visible bug fixes in chezmoi code
  * `docs:` for adding to or updating the documentation
  * `chore:` for user-irrelevant changes, e.g. fixing a typo, improving test/CI
    infrastructure, or updating dependencies

Commit message scopes are not allowed (e.g. `feat(cmd):` is invalid), and the
first letter after the prefix must be uppercase.

Example commit messages are:

    feat: Add shellQuote and shellQuoteList template funcs
    fix: Don't warn on secrets when re-adding encrypted files
    docs: Add push command to daily operations
    chore: Update dependencies

More examples can be found in the [commit history][history].

## Pull requests

All changes are made via pull requests. In your pull request, please ensure
that:

* All existing tests pass. You can ensure this by running `make test`.

* There are appropriate additional tests that demonstrate that your PR works as
  intended.

* The documentation is updated, if necessary. For new features you should add an
  entry in `assets/chezmoi.io/docs/user-guide/` and a complete description in
  `assets/chezmoi.io/docs/reference/`. See the [website developer guide][website]
  for instructions on how to build and view a local version of the documentation.
  By default, chezmoi will panic if a flag is undocumented or a long help is
  missing for a command. You can disable this panic during development by
  setting the environment variable `CHEZMOIDEV` to `ignoreflags=1,ignorehelp=1`.
  Once you have documented the command and its flags, run `make generate` to
  generate the embedded documentation.

* All generated files are up to date. You can ensure this by running `make
  generate` and including any modified files in your commit.

* The code is correctly formatted. You can ensure this by running `make format`.

* The code passes [`golangci-lint`][golangci-lint]. You can ensure this by
  running `make lint`.

* Commits are logically separate, with no merge or "fixup" commits.

* The branch applies cleanly to `master`.

[issue]: https://github.com/twpayne/chezmoi/issues/new/choose
[pr]: https://help.github.com/en/articles/creating-a-pull-request
[rationale]: https://dave.cheney.net/2019/02/18/talk-then-code
[golangci-lint]: https://github.com/golangci/golangci-lint
[commits]: https://www.conventionalcommits.org/en/v1.0.0/
[history]: https://github.com/twpayne/chezmoi/commits/master/
[website]: /developer-guide/website.md
[coc]: https://github.com/twpayne/chezmoi/blob/master/.github/CODE_OF_CONDUCT.md
[release-history-generator]: https://github.com/twpayne/chezmoi/blob/master/assets/chezmoi.io/docs/reference/release-history.md.tmpl
