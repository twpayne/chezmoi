# Contributing

`chezmoi` development happens on Github. When contributing, please first [open
an issue](https://github.com/twpayne/dotfiles/issues/new) to discuss the change
that you want to make. Bug reports and documentation improvements are
particularly welcome.

All changes are made via pull requests. In your pull request, please make sure
that:

* The commit messages match `chezmoi`'s convention, specifically that they being
  with a capitalized verb in the imperative and give a short description of what
  the commit does. Detailed information or justification can be optionally
  included in the body of the commit message.

* Commits are logically separate, with no merge or "fixup" commits.

* All tests pass.

* The branch applies cleanly to `master`.

## Release management

Releases are managed with [goreleaser](https://goreleaser.com/).

To create a new release, push a tag, eg:

    git tag -a v0.1.0 -m "First release"
    git push origin v0.1.0

To run a local "snapshot" build without publishing:

    TRAVIS_BUILD_NUMBER=1 goreleaser --snapshot --rm-dist --debug --skip-publish
