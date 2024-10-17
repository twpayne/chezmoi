# `upgrade`

Upgrade chezmoi by downloading and installing the latest released version. This
will call the GitHub API to determine if there is a new version of chezmoi
available, and if so, download and attempt to install it in the same way as
chezmoi was previously installed.

If the any of the `$CHEZMOI_GITHUB_ACCESS_TOKEN`, `$CHEZMOI_GITHUB_TOKEN`,
`$GITHUB_ACCESS_TOKEN`, or `$GITHUB_TOKEN` environment variables are set, then
the first value found will be used to authenticate requests to the GitHub API,
otherwise unauthenticated requests are used which are subject to stricter [rate
limiting](https://developer.github.com/v3/#rate-limiting). Unauthenticated
requests should be sufficient for most cases.

!!! warning

    If you installed chezmoi using a package manager, the `upgrade` command
    might have been removed by the package maintainer.

## `--executable` *filename*

Set name of executable to replace.

## `--method` *method*

Set upgrade method.

| Methods                | Description                                                                        |
| ---------------------- | ---------------------------------------------------------------------------------- |
| `brew-upgrade`         | Run `brew upgrade chezmoi`.                                                        |
| `replace-executable`   | Download the latest released executable from Github.                               |
| `snap-refresh`         | Run `snap refresh chezmoi`.                                                        |
| `sudo-upgrade-package` | Same as `upgrade-package` but use `sudo`.                                          |
| `upgrade-package`      | Download and install `.apk`, `.deb` or `.rpm` package. Run `pacman` on Arch Linux. |
