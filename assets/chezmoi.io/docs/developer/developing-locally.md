# Developing locally

chezmoi is written in [Go](https://golang.org) and development happens on
[GitHub](https://github.com). chezmoi is a standard Go project, using standard
Go tooling. chezmoi requires Go 1.19 or later.

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
