# chezmoi install guide

<!--- toc --->
* [One-line binary install](#one-line-binary-install)
* [One-line package install](#one-line-package-install)
* [Pre-built Linux packages](#pre-built-linux-packages)
* [Pre-built binaries](#pre-built-binaries)
* [All pre-built Linux packages and binaries](#all-pre-built-linux-packages-and-binaries)
* [From source](#from-source)

## One-line binary install

Install the correct binary for your operating system and architecture in `./bin`
with a single command:

```console
$ sh -c "$(curl -fsLS git.io/chezmoi)"
```

Or, if you have `wget` instead of `curl`:

```console
$ sh -c "$(wget -qO- git.io/chezmoi)"
```

If you already have a dotfiles repo using chezmoi on GitHub at
`https://github.com/<github-username>/dotfiles` then you can install chezmoi and
your dotfiles with the single command:

```console
$ sh -c "$(curl -fsLS git.io/chezmoi)" -- init --apply <github-username>
```

Or on systems with Powershell, you can use one of the following command:

```
# To install in ./bin
(iwr -UseBasicParsing https://git.io/chezmoi.ps1).Content | powershell -c -

# To install in another location
'$params = "-BinDir ~/other"', (iwr https://git.io/chezmoi.ps1).Content | powershell -c -

# For information about other options, run
'$params = "-?"', (iwr https://git.io/chezmoi.ps1).Content | powershell -c -
```

## One-line package install

Install chezmoi with a single command.

| OS           | Method     | Command                                                                                     |
| ------------ | ---------- | ------------------------------------------------------------------------------------------- |
| Linux        | snap       | `snap install chezmoi --classic`                                                            |
| Linux        | Linuxbrew  | `brew install chezmoi`                                                                      |
| Alpine Linux | apk        | `apk add chezmoi`                                                                           |
| Arch Linux   | pacman     | `pacman -S chezmoi`                                                                         |
| Guix Linux   | guix       | `guix install chezmoi`                                                                      |
| NixOS Linux  | nix-env    | `nix-env -i chezmoi`                                                                        |
| Void Linux   | xbps       | `xbps-install -S chezmoi`                                                                   |
| macOS        | Homebrew   | `brew install chezmoi`                                                                      |
| macOS        | MacPorts   | `sudo port install chezmoi`                                                                 |
| Windows      | Scoop      | `scoop bucket add twpayne https://github.com/twpayne/scoop-bucket && scoop install chezmoi` |
| Windows      | Chocolatey | `choco install chezmoi`                                                                     |
| FreeBSD      | pkg        | `pkg install chezmoi`                                                                       |

## Pre-built Linux packages

Download a package for your operating system and architecture and install it
with your package manager.

| Distribution | Architectures                                             | Package                                                     |
| ------------ | --------------------------------------------------------- | ----------------------------------------------------------- |
| Alpine       | `386`, `amd64`, `arm64`, `arm`, `ppc64`, `ppc64le`        | [`apk`](https://github.com/twpayne/chezmoi/releases/latest) |
| Debian       | `amd64`, `arm64`, `armel`, `i386`, `ppc64`, `ppc64le`     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest) |
| RedHat       | `aarch64`, `armhfp`, `i686`, `ppc64`, `ppc64le`, `x86_64` | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenSUSE     | `aarch64`, `armhfp`, `i686`, `ppc64`, `ppc64le`, `x86_64` | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest) |
| Ubuntu       | `amd64`, `arm64`, `armel`, `i386`, `ppc64`, `ppc64le`     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest) |

## Pre-built binaries

Download an archive for your operating system containing a pre-built binary,
documentation, and shell completions.

| OS         | Architectures                                       | Archive                                                        |
| ---------- | --------------------------------------------------- | -------------------------------------------------------------- |
| FreeBSD    | `amd64`, `arm`, `arm64`, `i386`                     | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Linux      | `amd64`, `arm`, `arm64`, `i386`, `ppc64`, `ppc64le` | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| macOS      | `amd64`, `arm64`                                    | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenBSD    | `amd64`, `arm`, `arm64`, `i386`                     | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Windows    | `amd64`, `i386`                                     | [`zip`](https://github.com/twpayne/chezmoi/releases/latest)    |

## All pre-built Linux packages and binaries

All pre-built binaries and packages can be found on the [chezmoi GitHub releases
page](https://github.com/twpayne/chezmoi/releases/latest).

## From source

Download, build, and install chezmoi for your system:

```console
$ go install github.com/twpayne/chezmoi@latest
```

Building chezmoi requires Go 1.17 or later.
