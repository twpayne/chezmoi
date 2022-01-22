# Releases

Releases are managed with [`goreleaser`](https://goreleaser.com/).

## Testing

To build a test release, without publishing, (Linux only) run:

```console
$ make test-release
```

## Publishing

Publish a new release by creating and pushing a tag, for example:

```console
$ git tag v1.2.3
$ git push --tags
```

This triggers a [GitHub Action](https://github.com/twpayne/chezmoi/actions)
that builds and publishes archives, packages, and snaps, and creates a new
[GitHub Release](https://github.com/twpayne/chezmoi/releases).

!!! note

    Publishing [Snaps](https://snapcraft.io/) requires a `SNAPCRAFT_LOGIN`
    [repository
    secret](https://github.com/twpayne/chezmoi/settings/secrets/actions).
    Snapcraft logins periodically expire. Create a new snapcraft login by
    running:

    ```console
    $ snapcraft export-login --snaps=chezmoi --channels=stable,candidate,beta,edge --acls=package_upload -
    ```

!!! note

    [brew](https://brew.sh/) automation will automatically detect new releases
    of chezmoi within a few hours and open a pull request in
    [github.com/Homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core)
    to bump the version.

    If needed, the pull request can be created with:

    ```console
    $ brew bump-formula-pr --tag=v1.2.3 chezmoi
    ```

!!! note

    chezmoi is in [Scoop](https://scoop.sh/)'s Main bucket. Scoop's automation
    will automatically detect new releases within a few hours.
