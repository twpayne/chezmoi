# Packaging

If you're packaging chezmoi for an operating system or distribution:

## Build

chezmoi has no build dependencies other than the standard Go toolchain.

### Dependencies

chezmoi has no runtime dependencies, but is usually used with [git][git], so
many packagers choose to make `git` an install dependency or recommended
package.

### Versioning

Please set the version number, git commit, and build time in the binary. This
greatly assists debugging when end users report problems or ask for help. You
can do this by passing the following flags to `go build`:

```text
-ldflags "-X main.version=$VERSION
          -X main.commit=$COMMIT
          -X main.date=$DATE
          -X main.builtBy=$BUILT_BY"
```

`$VERSION` should be the chezmoi version, e.g. `2.50.1`. Any `v` prefix is
optional and will be stripped, so you can pass the git tag in directly.

!!! hint

    The command `git describe --abbrev=0 --tags` will return a suitable value
    for `$VERSION`.

`$COMMIT` should be the full git commit hash at which chezmoi is built, e.g.
`4d678ce6850c9d81c7ab2fe0d8f20c1547688b91`.

!!! hint

    The `internal/cmds/generate-commit/main.go` script will return a suitable value
    for `$COMMIT`.

    You can run it with `go tool generate-commit`.

!!! hint

    The source archive contains a file called `COMMIT` containing the commit
    hash.

`$DATE` should be the date of the build as a UNIX timestamp or in RFC3339
format.

!!! hint

    The command `git show -s --format=%ct HEAD` returns the UNIX timestamp of
    the last commit, e.g. `1636668628`.

    The command `date -u +%Y-%m-%dT%H:%M:%SZ` returns the current time in
    RFC3339 format, e.g. `2019-11-23T18:29:25Z`.

`$BUILT_BY` should be a string indicating what system was used to build the
binary. Typically it should be the name of your packaging system, e.g.
`homebrew`.

### cgo

Please enable cgo, if possible. chezmoi can be built and run without cgo, but
the `.chezmoi.username` and `.chezmoi.group` template variables may not be set
correctly on some systems.

### Disabling the `upgrade` command

chezmoi includes an [`upgrade` command][upgrade-command] which attempts to
self-upgrade. If you're packaging chezmoi for a distribution that includes its
own upgrade mechanism, this should be disabled.

You can remove the `upgrade` command completely by building chezmoi with the
`noupgrade` build tag, for example

    go build -tags=noupgrade .

## Shell completions

chezmoi includes shell completions in the [`completions` directory][completions]. Please
include these in the package and install them in the shell-appropriate
directory, if possible.

## Install instructions

If the instructions for installing chezmoi in chezmoi's [install guide][guide]
are absent or incorrect, please open an issue or submit a PR to correct them.

[completions]: https://github.com/twpayne/chezmoi/tree/master/completions
[git]: https://git-scm.com/
[go.mod]: https://github.com/twpayne/chezmoi/blob/master/go.mod
[guide]: /install.md
[upgrade-command]: /reference/commands/upgrade.md
