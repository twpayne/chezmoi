# Releases

Releases are managed with [`goreleaser`][goreleaser].

## Testing

To build a test release, without publishing, (Ubuntu Linux only) first ensure
that the `musl-tools` and `snapcraft` packages are installed:

```sh
sudo apt-get install musl-tools snapcraft
```

Then run:

```sh
make test-release
```

## Publishing

Publish a new release by creating and pushing a tag, for example:

```sh
git tag v1.2.3
git push --tags
```

This triggers a [GitHub Action][gha] that builds and publishes archives,
packages, and snaps, creates a new [GitHub Release][release], and deploys the
[website][website].

### Snaps

Publishing [Snaps][snaps] requires a `SNAPCRAFT_STORE_CREDENTIALS` [repository
secret][secret].

Snapcraft store credentials periodically expire. Create new snapcraft store
credentials by running:

```sh
snapcraft export-login --snaps=chezmoi --channels=stable,candidate,beta,edge --acls=package_upload -
```

### Homebrew

[Homebrew][homebrew] automation will automatically detect new releases of chezmoi within
a few hours and open a pull request in
[github.com/Homebrew/homebrew-core][homebrew-core] to bump the version.

If needed, the pull request can be created with:

```sh
brew bump-formula-pr --tag=v1.2.3 chezmoi
```

### Scoop

chezmoi is in [Scoop][scoop]'s Main bucket. Scoop's automation will
automatically detect new releases within a few hours.

## Signing

chezmoi uses [GoReleaser's support for signing][signing] to sign the checksums
of its release assets with [cosign][cosign].

Details:

* The cosign private key was generated with cosign v1.12.1 on a private
  recently-installed Ubuntu 22.04.1 system with a single user and all available
  updates applied.

* The private key uses a long (more than 32 character) password generated
  locally by a password manager.

* The password-protected private key is stored in chezmoi's public GitHub repo.

* The private key's password is stored as a [GitHub Actions secret][gha-secret]
  and only available to the `release` step of `release` job of the `main`
  workflow.

* The cosign public key is included in the release assets and also uploaded to
  [`https://chezmoi.io/cosign.pub`][pubkey]. Since
  [`https://chezmoi.io`][website] is served by [GitHub pages][pages], it
  probably has equivalent security to [chezmoi's GitHub Releases page][release],
  which is also managed by GitHub.

[goreleaser]: https://goreleaser.com/
[gha]: https://github.com/twpayne/chezmoi/actions
[release]: https://github.com/twpayne/chezmoi/releases
[website]: https://chezmoi.io
[snaps]: https://snapcraft.io/
[secret]: https://github.com/twpayne/chezmoi/settings/secrets/actions
[homebrew]: https://brew.sh/
[homebrew-core]: https://github.com/Homebrew/homebrew-core
[scoop]: https://scoop.sh/
[signing]: https://goreleaser.com/customization/sign/
[cosign]: https://github.com/sigstore/cosign
[gha-secret]: https://docs.github.com/en/actions/security-guides/encrypted-secrets
[pubkey]: https://chezmoi.io/cosign.pub
[pages]: https://pages.github.com/
