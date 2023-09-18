# Developing locally

chezmoi is written in [Go](https://golang.org) and development happens on
[GitHub](https://github.com). chezmoi is a standard Go project, using standard
Go tooling. chezmoi requires Go 1.20 or later.

Checkout chezmoi:

```console
$ git clone https://github.com/twpayne/chezmoi.git
$ cd chezmoi
```

Build chezmoi:

```console
$ go build
```

!!! hint

    If you try to build chezmoi with an unsupported version of Go you will get
    the error:

    ```
    package github.com/twpayne/chezmoi/v2: build constraints exclude all Go files in /home/twp/src/github.com/twpayne/chezmoi
    ```

Run all tests:

```console
$ go test ./...
```

chezmoi's tests include integration tests with other software. If the other
software is not found in `$PATH` the tests will be skipped. Running the full
set of tests requires `age`, `base64`, `bash`, `gpg`, `perl`, `python3`,
`ruby`, `sed`, `sha256sum`, `unzip`, `xz`, `zip`, and `zstd`.

Run chezmoi:

```console
$ go run .
```

Run a set of smoketests, including cross-compilation, tests, and linting:

```console
$ make smoketest
```

!!! hint

    If you use `fish` as your primary shell, you may get warnings from Fish
    during tests:

    ```
    error: can not save history
    warning-path: Unable to locate data directory derived from $HOME: '/home/user/.local/share/fish'.
    warning-path: The error was 'Operation not supported'.
    warning-path: Please set $HOME to a directory where you have write access.
    ```

    These can be avoided with by running tests with `SHELL=bash` or `SHELL=zsh`:

    ```console
    $ SHELL=bash make test
    $ SHELL=zsh make smoketest
    $ SHELL=bash go test ./...
    ```
