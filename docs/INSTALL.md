# chezmoi Install Guide

<!--- toc --->
* [One-line binary install](#one-line-binary-install)
* [One-line package install](#one-line-package-install)
* [Pre-built Linux packages](#pre-built-linux-packages)
* [Pre-built binaries](#pre-built-binaries)
* [All pre-built Linux packages and binaries](#all-pre-built-linux-packages-and-binaries)
* [From source](#from-source)
* [Upgrading](#upgrading)

## One-line binary install

Install the correct binary for your operating system and architecture in `./bin`
with a single command.

    curl -sfL https://git.io/chezmoi | sh

## One-line package install

Install chezmoi with a single command.

| OS           | Method     | Command                                                                                     |
| ------------ | ---------- | ------------------------------------------------------------------------------------------- |
| Linux        | snap       | `snap install chezmoi --classic`                                                            |
| Linux        | Linuxbrew  | `brew install twpayne/taps/chezmoi`                                                         |
| Alpine Linux | apk        | `apk add chezmoi`                                                                           |
| Arch Linux   | pacman     | `pacman -S chezmoi`                                                                         |
| NixOS Linux  | nix-env    | `nix-env -i chezmoi`                                                                        |
| macOS        | Homebrew   | `brew install twpayne/taps/chezmoi`                                                         |
| Windows      | Scoop      | `scoop bucket add twpayne https://github.com/twpayne/scoop-bucket && scoop install chezmoi` |

## Pre-built Linux packages

Download a package for your operating system and architecture and install it
with your package manager.

| Distribution | Architectures                                             | Package                                                                   |
| ------------ | --------------------------------------------------------- | ------------------------------------------------------------------------- |
| Debian       | `amd64`, `arm64`, `armel`, `i386`, `ppc64`, `ppc64le`     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest)               |
| RedHat       | `aarch64`, `armhfp`, `i686`, `ppc64`, `ppc64le`, `x86_64` | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest)               |
| OpenSUSE     | `aarch64`, `armhfp`, `i686`, `ppc64`, `ppc64le`, `x86_64` | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest)               |
| Ubuntu       | `amd64`, `arm64`, `armel`, `i386`, `ppc64`, `ppc64le`     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest)               |

## Pre-built binaries

Download an archive for your operating system containing a pre-built binary,
documentation, and shell completions.

| OS         | Architectures                                       | Archive                                                        |
| ---------- | --------------------------------------------------- | -------------------------------------------------------------- |
| FreeBSD    | `amd64`, `arm`, `i386`                              | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Linux      | `amd64`, `arm`, `arm64`, `i386`, `ppc64`, `ppc64le` | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| macOS      | `amd64`                                             | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenBSD    | `amd64`, `arm`, `i386`                              | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Windows    | `amd64`, `i386`                                     | [`zip`](https://github.com/twpayne/chezmoi/releases/latest)    |

## All pre-built Linux packages and binaries

All pre-built binaries and packages can be found on the [chezmoi GitHub releases
page](https://github.com/twpayne/chezmoi/releases/latest).

## From source

Download, build, and install chezmoi for your system:

    cd $(mktemp -d)
    git clone --depth=1 https://github.com/twpayne/chezmoi.git
    cd chezmoi
    go install

Building chezmoi requires Go 1.13 or later.

## Upgrading

If you have installed a pre-built binary of chezmoi, you can upgrade it to the
latest release with:

    chezmoi upgrade

This will re-use whichever mechanism you used to install chezmoi to install the
latest release.
