# Releases

Releases are managed with [`goreleaser`](https://goreleaser.com/).

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

This triggers a [GitHub Action](https://github.com/twpayne/chezmoi/actions)
that builds and publishes archives, packages, and snaps, creates a new [GitHub
Release](https://github.com/twpayne/chezmoi/releases), and deploys the
[website](https://chezmoi.io).

!!! note

    Publishing [Snaps](https://snapcraft.io/) requires a
    `SNAPCRAFT_STORE_CREDENTIALS` [repository
    secret](https://github.com/twpayne/chezmoi/settings/secrets/actions).

    Snapcraft store credentials periodically expire. Create new snapcraft store
    credentials by running:

    ```sh
    snapcraft export-login --snaps=chezmoi --channels=stable,candidate,beta,edge --acls=package_upload -
    ```

!!! note

    [brew](https://brew.sh/) automation will automatically detect new releases
    of chezmoi within a few hours and open a pull request in
    [github.com/Homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core)
    to bump the version.

    If needed, the pull request can be created with:

    ```sh
    brew bump-formula-pr --tag=v1.2.3 chezmoi
    ```

!!! note

    chezmoi is in [Scoop](https://scoop.sh/)'s Main bucket. Scoop's automation
    will automatically detect new releases within a few hours.

## Signing

chezmoi uses [GoReleaser's support for
signing](https://goreleaser.com/customization/sign/) to sign the checksums of
its release assets with [cosign](https://github.com/sigstore/cosign).

Details:

* The cosign private key was generated with cosign v1.12.1 on a private
  recently-installed Ubuntu 22.04.1 system with a single user and all available
  updates applied.

* The private key uses a long (more than 32 character) password generated
  locally by a password manager.

* The password-protected private key is stored in chezmoi's public GitHub repo.

* The private key's password is stored as a [GitHub Actions
  secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
  and only available to the `release` step of `release` job of the `main`
  workflow.

* The cosign public key is included in the release assets and also uploaded to
  [`https://chezmoi.io/cosign.pub`](https://chezmoi.io/cosign.pub). Since
  [`https://chezmoi.io`](https://chezmoi.io) is served by [GitHub
  pages](https://pages.github.com/), it probably has equivalent security to
  [chezmoi's GitHub Releases
  page](https://github.com/twpayne/chezmoi/releases), which is also managed by
  GitHub.
