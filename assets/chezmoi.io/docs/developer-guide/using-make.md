# Building and installing with `make`

chezmoi can be built with GNU make, assuming you have the Go toolchain
installed.

Running `make` will build a `chezmoi` binary in the current directory for the
host OS and architecture. To embed version information in the binary and
control installation the following variables are available:

| Variable    | Example                | Purpose                                        |
| ----------- | ---------------------- | ---------------------------------------------- |
| `$VERSION`  | `v2.0.0`               | Set version                                    |
| `$COMMIT`   | `3895680a`...          | Set the git commit at which the code was built |
| `$DATE`     | `2019-11-23T18:29:25Z` | The time of the build                          |
| `$BUILT_BY` | `homebrew`             | The packaging system performing the build      |
| `$PREFIX`   | `/usr`                 | Installation prefix                            |
| `$DESTDIR`  | `install-root`         | Fake installation root                         |

Running `make install` will install the `chezmoi` binary in
`${DESTDIR}${PREFIX}/bin`.
