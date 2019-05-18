# Installation

One line install:

    curl -sfL https://git.io/chezmoi | sh

On macOS you can install chezmoi with [Homebrew](https://brew.sh):

    brew install twpayne/taps/chezmoi

On Linux distributions with [snap](https://snapcraft.io), you can install
chezmoi with:

    snap install chezmoi --classic

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

If you have Go installed you can install the latest version from `HEAD`:

    go get -u github.com/twpayne/chezmoi

Once chezmoi is installed, you can upgrade it to the latest release with:

    chezmoi upgrade