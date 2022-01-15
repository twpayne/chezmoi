# Install

## One-line binary install

Install the correct binary for your operating system and architecture in `./bin`
with a single command:

=== "curl"

    ```sh
    sh -c "$(curl -fsLS chezmoi.io/get)"
    ```

=== "wget"

    ```sh
    sh -c "$(wget -qO- chezmoi.io/get)"
    ```

=== "PowerShell"

    ```powershell
    (iwr -UseBasicParsing https://chezmoi.io/get.ps1).Content | powershell -c -
    ```

!!! hint

    If you already have a dotfiles repo using chezmoi on GitHub at
    `https://github.com/<github-username>/dotfiles` then you can install
    chezmoi and your dotfiles with the single command:

    ```sh
    sh -c "$(curl -fsLS chezmoi.io/get)" -- init --apply <github-username>
    ```

## One-line package install

Install chezmoi with your package manager with a single command:

=== "Linux"

    === "snap"

        ```sh
        snap install chezmoi --classic
        ```

    === "Linuxbrew"

        ```sh
        brew install chezmoi
        ```

    === "asdf"

        ```sh
        asdf plugin add chezmoi && asdf install chezmoi <version>
        ```

    === "Alpine"

        ```sh
        apk add chezmoi
        ```

    === "Arch"

        ```sh
        pacman -S chezmoi
        ```

    === "Guix"

        ```sh
        guix install chezmoi
        ```

    === "Nix / NixOS"

        ```sh
        nix-env -i chezmoi
        ```

    === "Void"

        ```sh
        xbps-install -S chezmoi
        ```

=== "macOS"

    === "Homebrew"

        ```sh
        brew install chezmoi
        ```

    === "MacPorts"

        ```sh
        port install chezmoi
        ```

    === "Nix"

        ```sh
        nix-env -i chezmoi
        ```

    === "asdf"

        ```sh
        asdf plugin add chezmoi && asdf install chezmoi <version>
        ```

=== "Windows"

    === "Chocolately"

        ```
        choco install chezmoi
        ```

    === "Scoop"

        ```
        scoop bucket add twpayne https://github.com/twpayne/scoop-bucket && scoop install chezmoi
        ```

=== "FreeBSD"

    ```sh
    pkg install chezmoi
    ```

=== "OpenIndiana"

    ```sh
    pkg install application/chezmoi
    ```

## Download a pre-built Linux package

Download a package for your operating system and architecture and install it
with your package manager.

| Distribution | Architectures                                             | Package                                                     |
| ------------ | --------------------------------------------------------- | ----------------------------------------------------------- |
| Alpine       | `386`, `amd64`, `arm64`, `arm`, `ppc64`, `ppc64le`        | [`apk`](https://github.com/twpayne/chezmoi/releases/latest) |
| Debian       | `amd64`, `arm64`, `armel`, `i386`, `ppc64`, `ppc64le`     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest) |
| RedHat       | `aarch64`, `armhfp`, `i686`, `ppc64`, `ppc64le`, `x86_64` | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenSUSE     | `aarch64`, `armhfp`, `i686`, `ppc64`, `ppc64le`, `x86_64` | [`rpm`](https://github.com/twpayne/chezmoi/releases/latest) |
| Ubuntu       | `amd64`, `arm64`, `armel`, `i386`, `ppc64`, `ppc64le`     | [`deb`](https://github.com/twpayne/chezmoi/releases/latest) |

## Download a pre-built binary

Download an archive for your operating system containing a pre-built binary and
shell completions.

| OS         | Architectures                                       | Archive                                                        |
| ---------- | --------------------------------------------------- | -------------------------------------------------------------- |
| FreeBSD    | `amd64`, `arm`, `arm64`, `i386`                     | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Illumos    | `amd64`                                             | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Linux      | `amd64`, `arm`, `arm64`, `i386`, `ppc64`, `ppc64le` | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| macOS      | `amd64`, `arm64`                                    | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenBSD    | `amd64`, `arm`, `arm64`, `i386`                     | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Solaris    | `amd64`                                             | [`tar.gz`](https://github.com/twpayne/chezmoi/releases/latest) |
| Windows    | `amd64`, `arm`, `i386`                              | [`zip`](https://github.com/twpayne/chezmoi/releases/latest)    |

## Install from source

Download, build, and install chezmoi for your system with Go 1.16 or later:

```console
$ git clone https://github.com/twpayne/chezmoi.git
$ cd chezmoi
$ make install
```
