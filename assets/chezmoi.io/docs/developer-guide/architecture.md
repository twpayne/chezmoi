# Architecture

This document gives a high-level overview of chezmoi's source code for anyone
interested in contributing to chezmoi.

You can generate Go documentation for chezmoi's source code with `go doc`, for
example:

```sh
go doc -all -u github.com/twpayne/chezmoi/internal/chezmoi
```

You can also [browse chezmoi's generated documentation online][go-docs].

## Directory structure

The important directories in chezmoi are:

| Directory                        | Contents                                                                                                                                            |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `assets/chezmoi.io/docs/`        | The documentation single source of truth. Help text, examples, and the [chezmoi.io][website] website are generated from the files in this directory |
| `internal/chezmoi/`              | chezmoi's core functionality                                                                                                                        |
| `internal/cmd/`                  | Code for the `chezmoi` command                                                                                                                      |
| `internal/cmd/testdata/scripts/` | High-level tests of chezmoi's commands using [`testscript`][testscript]                                                                             |

## Key concepts

As described in the [reference manual][ref], chezmoi evaluates the source state
to compute a target state for the destination directory (typically your home
directory). It then compares the target state to the actual state of the
destination directory and performs any changes necessary to update the
destination directory to match the target state. These concepts are represented
directly in chezmoi's code.

chezmoi uses the generic term *entry* to describe something that it manages.
Entries can be files, directories, symlinks, scripts, amongst other things.

## `internal/chezmoi/` directory

All of chezmoi's interaction with the operating system is abstracted through the
`System` interface. A `System` includes functionality to read and write files
and directories and execute commands. chezmoi makes a distinction between
idempotent commands that can be run multiple times without modifying the
underlying system and arbitrary commands that may modify the underlying system.

The real underlying system is implemented via a `RealSystem` struct. Other
`System`s are composed on top of this to provide further functionality. For
example, the `--debug` flag is implemented by wrapping the `RealSystem` with a
`DebugSystem` that logs all calls to the underlying `RealSystem`. `--dry-run` is
implemented by wrapping the `RealSystem` with a `DryRunSystem` that allows reads
to pass through but silently discards all writes.

The `SourceState` struct represents a source state, including reading a source
state from the source directory, executing templates, applying the source state
(i.e. updating a `System` to match the desired source state), and adding more
entries to the source state.

Entries in the source state are abstracted by the `SourceStateEntry` interface
implemented by the `SourceStateFile` and `SourceStateDir` structs, as the source
state only consists of regular files and directories.

A `SourceStateFile` includes a `FileAttr` struct describing the attributes
parsed from its file name. Similarly, a `SourceStateDir` includes a `DirAttr`
struct describing the directory attributes parsed from a directory name.

`SourceStateEntry`s can compute their target state entries, i.e. what the
equivalent entry should be in the target state, abstracted by the
`TargetStateEntry` interface.

Actual target state entries include `TargetStateFile` structs, representing a
file with contents and permissions, `TargetStateDir` structs, representing a
directory, `TargetStateSymlink` for symlinks, `TargetStateRemove` for entries
that should be removed, and `TargetStateScript` for scripts that should be run.

The actual state of an entry in the target state is abstracted via the
`ActualStateEntry` interface, with `ActualStateAbsent`, `ActualStateDir`,
`ActualStateFile`, `ActualStateSymlink` structs implementing this interface.

Finally, an `EntryState` struct represents a serialization of an
`ActualEntryState` for storage in and retrieval from chezmoi's persistent state.
It stores a SHA256 of the entry's contents, rather than the full contents, to
avoid storing secrets in the persistent state.

With these concepts, chezmoi's apply command is effectively:

1. Read the source state from the source directory.

2. For each entry in the source state (`SourceStateEntry`), compute its
   `TargetStateEntry` and read its actual state in the destination state
   (`ActualStateEntry`).

3. If the `ActualStateEntry` is not equivalent to the `TargetStateEntry` then
   apply the minimal set of changes to the `ActualStateEntry` so that they are
   equivalent.

Furthermore, chezmoi stores the `EntryState` of each entry that it writes in its
persistent state. chezmoi can then detect if a third party has updated a target
since chezmoi last wrote it by comparing the actual state entry in the target
state with the entry state in the persistent state.

## `internal/cmd/` directory

`internal/cmd/*cmd.go` files contain the code for each individual command.
`internal/cmd/*templatefuncs.go` files contain the template functions.

Commands are defined as methods on the `Config` struct. The `Config` struct is
large, containing all configuration values read from the config file, command
line arguments, and computed and cached values.

The `Config.persistentPreRunRootE` and `Config.persistentPostRunRootE` methods
set up and tear down state for individual commands based on the command's
`Annotations` field, which defines how the command interacts with the file
system and persistent state.

## Path handling

chezmoi uses separate types for absolute paths (`AbsPath`) and relative paths
(`RelPath`) to avoid errors where paths are combined (e.g. joining two absolute
paths is an error). The type `SourceRelPath` is a relative path within the
source directory and handles file and directory attributes.

Internally, chezmoi normalizes all paths to use forward slashes with an optional
upper-cased Windows volume so they can be compared with string comparisons.
Paths read from the user may include tilde (`~`) to represent the user's home
directory, use forward or backward slashes, and are treated as external paths
(`ExtPath`). These are normalized to absolute paths. chezmoi is case-sensitive
internally and makes no attempt to handle case-insensitive or case-preserving
file systems.

## Persistent state

Persistent state is treated as a two-level key-value store with the
pseudo-structure `map[Bucket]map[Key]Value`, where `Bucket`, `Key`, and `Value`
are all `[]byte`s. The `PersistentState` interface defines interaction with
them. Sometimes temporary persistent states are used. For example, in dry run
mode (`--dry-run`) the actual persistent state is copied into a temporary
persistent state in memory which remembers writes but does not persist them to
disk.

## Encryption

Encryption tools are abstracted by the `Encryption` interface that contains
methods of encrypting and decrypting files and `[]byte`s. Implementations are
the `AGEEncryption` and `GPGEncryption` structs. A `DebugEncryption` struct
wraps an `Encryption` interface and logs the methods called.

## `run_once_` and `run_onchange_` scripts

The execution of a `run_once_` script is recorded by storing the SHA256 of its
contents in the `scriptState` bucket in the persistent state. On future
invocations the script is only run if no matching contents SHA256 is found in
the persistent state.

The execution of a `run_onchange_` script is recorded by storing its target name
in the `entryState` bucket along with its contents SHA256 sum. On future
invocations the script is only run if its contents SHA256 sum has changed, and
its contents SHA256 sum is then updated in the persistent state.

## Testing

chezmoi has a mix of unit, integration, and end-to-end tests. Unit and
integration tests use the [`github.com/alecthomas/assert/v2`][assert] framework.
End-to-end tests use [`github.com/rogpeppe/go-internal/testscript`][testscript]
with the test scripts themselves in
`internal/cmd/testdata/scripts/$TEST_NAME.txtar`.

You can run individual end-to-end tests with

```sh
go test ./internal/cmd -run=TestScript/$TEST_NAME
```

Tests should, if at all possible, run unmodified on all operating systems tested
in CI (Linux, macOS, Windows, and FreeBSD). Windows will sometimes need special
handling due to its path separator and lack of POSIX-style file permissions.

[go-docs]: https://pkg.go.dev/github.com/twpayne/chezmoi
[website]: https://chezmoi.io
[testscript]: https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript
[ref]: /reference/concepts.md
[assert]: https://pkg.go.dev/github.com/alecthomas/assert
