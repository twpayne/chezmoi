# Installation

One line install:

    curl -sfL https://install.goreleaser.com/github.com/twpayne/chezmoi.sh | sh

Pre-built packages and binaries:

| OS         | Architectures                                 | Package location                                                     |
| ---------- | --------------------------------------------- | -------------------------------------------------------------------- |
| Arch Linux | i686, x86_64                                  | [aur package](https://aur.archlinux.org/packages/chezmoi)            |
| Debian     | amd64, arm64, armel, i386, ppc64, ppc64le     | [deb package](https://github.com/twpayne/chezmoi/releases/latest)    |
| RedHat     | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [rpm package](https://github.com/twpayne/chezmoi/releases/latest)    |
| OpenSUSE   | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [rpm package](https://github.com/twpayne/chezmoi/releases/latest)    |
| FreeBSD    | amd64, arm, i386                              | [tar.gz package](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenBSD    | amd64, arm, i386                              | [tar.gz package](https://github.com/twpayne/chezmoi/releases/latest) |
| Linux      | amd64, arm, arm64, i386, ppc64, ppc64le       | [tar.gz package](https://github.com/twpayne/chezmoi/releases/latest) |

On macOS you can install chezmoi with Homebrew:

    brew install twpayne/taps/chezmoi

If you have Go installed you can install the latest version from `HEAD`:

    go get -u github.com/twpayne/chezmoi
