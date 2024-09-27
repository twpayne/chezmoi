# Developer

chezmoi is written in [Go](https://golang.org) and development happens on
[GitHub](https://github.com). chezmoi is a standard Go project, using standard
Go tooling. chezmoi requires Go 1.22 or later.

Checkout chezmoi:

```console
$ git clone https://github.com/twpayne/chezmoi.git
$ cd chezmoi
```

Build chezmoi:

```console
$ go build
```

Run all tests:

```console
$ go test ./...
```

chezmoi's tests include integration tests with other software. If the other
software is not found in `$PATH` the tests will be skipped. Running the full set
of tests requires `age`, `base64`, `bash`, `bzip2`, `git`, `gpg`, `gzip`,
`perl`, `python3`, `rage`, `ruby`, `sed`, `sha256sum`, `tr`, `true`, `unzip`,
`xz`, `zip`, and `zstd`.

Run chezmoi:

```console
$ go run .
```

Run a set of smoke tests, including cross-compilation, tests, and linting:

```console
$ make smoke-test
```

Test building chezmoi for all architectures:

```console
$ make test-release
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
    $ SHELL=zsh make smoke-test
    $ SHELL=bash go test ./...
    ```
