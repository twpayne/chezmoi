# chezmoi Install Guide

* [One-line binary install](#one-line-binary-install)
* [One-line package install](#one-line-package-install)
* [Pre-built Linux packages](#pre-built-linux-packages)
* [Pre-built binaries](#pre-built-binaries)
* [From source](#from-source)
* [Upgrading](#upgrading)

## One-line binary install

Install the correct binary for your operating system and architecture with a
single command.

    curl -sfL https://git.io/chezmoi | sh

## One-line package install

Install chezmoi with a single command.

| OS    | Distribution | Method   | Command                             |
| ----- | ------------ | ---------| ----------------------------------- |
| Linux | -            | snap     | `snap install chezmoi --classic`    |
| Linux | Arch         | pacman   | `pacman -S chezmoi`                |
| macOS | -            | Homebrew | `brew install twpayne/taps/chezmoi` |

## Pre-built Linux packages

Download a package for your operating system and architecture and install it
with your package manager.

| OS         | Architectures                                 | Package                                                                   |
| ---------- | --------------------------------------------- | ------------------------------------------------------------------------- |
| Debian     | amd64, arm64, armel, i386, ppc64, ppc64le     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest)               |
| RedHat     | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest)               |
| OpenSUSE   | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest)               |
| Ubuntu     | amd64, arm64, armel, i386, ppc64, ppc64le     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest)               |

## Pre-built binaries

Download a tarball for your operating system containing a pre-built binary and
documentation.

| OS         | Architectures                                 | Tarball                                                        |
| ---------- | --------------------------------------------- | -------------------------------------------------------------- |
| FreeBSD    | amd64, arm, i386                              | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Linux      | amd64, arm, arm64, i386, ppc64, ppc64le       | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| macOS      | amd64                                         | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenBSD    | amd64, arm, i386                              | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |

## From source

Download, build, and install chezmoi for your system:

    cd $(mktemp -d) && go get -u github.com/twpayne/chezmoi

## Upgrading

Once chezmoi is installed, you can upgrade it to the latest release with:

    chezmoi upgrade

This will re-use whichever mechanism you used to install chezmoi to install the
latest release.
