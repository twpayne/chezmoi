# Installation

* [macOS with Homebrew](#macos-with-homebrew)
* [One-line binary install](#one-line-binary-install)
* [Pre-built packages and binaries](#pre-built-packages-and-binaries)
* [From source](#from-source)
* [Upgrading](#upgrading)

## macOS with Homebrew

    brew install twpayne/taps/chezmoi

## One-line binary install

    curl -sfL https://git.io/chezmoi | sh

## Pre-built packages and binaries

| OS         | Architectures                                 | Package                                                        |
| ---------- | --------------------------------------------- | -------------------------------------------------------------- |
| Arch Linux | i686, x86_64                                  | [`aur`](https://aur.archlinux.org/packages/chezmoi)            |
| Debian     | amd64, arm64, armel, i386, ppc64, ppc64le     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest)    |
| RedHat     | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest)    |
| OpenSUSE   | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest)    |
| FreeBSD    | amd64, arm, i386                              | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenBSD    | amd64, arm, i386                              | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Linux      | amd64, arm, arm64, i386, ppc64, ppc64le       | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |

## From source

    cd $(mktemp -d) && go get -u github.com/twpayne/chezmoi

## Upgrading

Once chezmoi is installed, you can upgrade it to the latest release with:

    chezmoi upgrade

This will re-use whichever mechanism you used to install chezmoi to install the
latest release.