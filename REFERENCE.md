# `chezmoi` Reference Manual

Manage your dotfiles securely across multiple machines.



## Global command line flags

Command line flags override any values set in the configuration file.


### `-c`, `--config` *filename*

Read the configuration from *filename*.


### `-D`, `--destination` *directory*

Use *directory* as the destination directory.


### `-n`, `--dry-run`

Set dry run mode. In dry run mode, the destination directory is never modified.
This is most useful in combination with the `-v` (verbose) flag to print
changes that would be made without making them.


### `-h`, `--help`

Print help.


### `-S`, `--source` *directory*

Use *directory* as the source directory.


### `-u`, `--umask`

Set the umask.


### `-v`, `--verbose`

Set verbose mode. In verbose mode, `chezmoi` prints the changes that it is
making as approximate shell commands, and any differences in files between the
target state and the destination set are printed as unified diffs.


### `--version`

Print the version of `chezmoi`, the commit at which it was built, and the build
timestamp.



## Configuration file

`chezmoi` searches for its configuration file according to the [XDG Base
Directory
Specification](https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html)
and supports all formats supported by
[`github.com/spf13/viper`](https://github.com/spf13/viper), namely JSON, TOML,
YAML, macOS property file format, and HCL. The basename of the config file is
`chezmoi`, and the first config file found is used.


### Configuration variables

The following configuration variables are available:

| Variable          | Type    | Default value             | Description                   |
| ----------------- | ------- | ------------------------- | ----------------------------- |
| `data`            | any     | none                      | Template data                 |
| `sourceDir`       | string  | `~/.config/share/chezmoi` | Source directory              |
| `targetDir`       | string  | `~`                       | Target directory              |
| `umask`           | integer | from system               | Umask                         |
| `dryRun`          | boolean | `false`                   | Dry run mode                  |
| `verbose`         | boolean | `false`                   | Verbose mode                  |
| `sourceVCSCommand | string  | `git`                     | Source version control system |

In addition, a number of secret manager integrations add configuration
variables. These are documented in the secret manager section.



## Targets

FIXME document
absolute paths
files, directories, and symlinks
attributes
move from "under the hood" section



## Commands


### `add` *targets*

Add *targets* to the source state. If any target is already in the source
state, then its source state is replaced with its current state in the
destination directory. The `add` command accepts additional flags:

#### `-e`, `--empty`

Set the `empty` attribute on added files.

#### `-x`, `--exact`

Set the `exact` attribute on added directories.

#### `-p`, `--prompt`

Interactively prompt before adding each file.

#### `-r`, `--recursive`

Recursively add all files, directories, and symlinks.

#### `-T`, `--template`

Set the `template` attribute on added files and symlinks. In addition,
`chezmoi` attempts to automatically generate the template by replacing any
template data values with the equivalent template data keys. Longer
subsitutions occur before shorter ones.


### `apply` [*targets*]

Ensure that *targets* are in the target state. If no targets are specified, the
state of all targets are ensured.


### `archive`

Write a tar archive of the target state to stdout. This can be piped into `tar`
to inspect the target state, for example:

    $ chezmoi archive | tar tvf -


### `cat` [targets]

Write the target state of *targets*  to stdout. *targets* must be files or
symlinks. For files, the target file contents are written. For symlinks, the
target target is written.


### `cd`

Launch a shell in the source directory.


### `chattr` attributes *targets*

Change the attributes of *targets*. *attributes* specifies which attributes to
modify. Add attributes by specifying them or their abbreviations directly,
optionally prefixed by `+`. Remove attributes by prefixing them or their
attributes with `no` or a `-`. The available attributes and their abbreviations
are:

| Attribute    | Abbreviation |
| ------------ | ------------ |
| `empty`      | `e`          |
| `exact`      | none         |
| `executable` | `x`          |
| `private`    | `p`          |
| `template`   | `t`          |

Multiple attributes modifications may be specified by separating them with `,`.

Examples:

    $ chezmoi chattr template ~/.bashrc
    $ chezmoi chattr noempty ~/.profile


### `data`

Write the computed template data in JSON format to stdout. The `data` command
accepts additional flags:

#### `-f`, `--format` *format*

Print the computed template data in the given format. The accepted formats are
`json` (JSON) and `yaml` (YAML).


### `diff` *targets*

Print the approximate shell commands required to ensure that *targets* in the
destination directory match the target state. If no targets are specifed, print
the commands required for all targets. It is equivalent to `chezmoi apply
--dry-run --verbose`.


### `doctor`

Check for potential problems.


### `dump` *targets*

Dump the target state in JSON format. If no targets are specified, then the
entire target state. The `dump` command accepts additional arguments:

#### `-f`, `--format` *format*

Print the target state in the given format. The accepted formats are `json`
(JSON) and `yaml` (YAML).


### `edit` *targets*

Edit the source state of *targets*, which must be files or symlinks. The `edit`
command accepts additional arguments:

#### `-a`, `--apply`

Apply target immediately after editing.

#### `-d`, `--diff`

Print the difference between the target state and the actual state after
editing.

#### `-p`, `--prompt`

Prompt before applying each target.


### `edit-config`

Edit the configuration file.


### `forget` *targets*

Remove *targets* from the source state, i.e. stop managing them.


### `help` *command*

Print the help associated with *command*.


### `import` *filename*

FIXME document

#### `-x`, `--exact`

#### `-r`, `--remove-destination`

#### `--strip-components` 


### `init` [*repo*]

FIXME document


### `remove`, `rm` *targets*

Remove *targets* from both the source state and the destination directory.


### `secret`

Interact with a secret manager. See the "Secret managers" section for details.


### `source` [*args*]

Execute the source version control system in the source directory with *args*.
Note that any flags for the source version control system must be sepeated with
a `--` to stop `chezmoi` from reading them.

Examples:

    $ chezmoi source init
    $ chezmoi source add .
    $ chezmoi source commit -- -m "Initial commit"


### `source-path` [*targets*]

Print the path to each target's source state. If no targets are specified then
print the source directory.


### `update`

FIXME document



### `verify` [*targets*]

Verify that all *targets* match their target state. `chezmoi` exits with code 0
(success) if all targets match their target state, or 1 (failure) otherwise. If
no targets are specified then all targets are checked.



## Editor configuration

The `edit` and `edit-config` commands use the editor specified by the `VISUAL`
environment variable, the `EDITOR` environment variable, or `vi`, whichever is
specified first.



## Umask

FIXME document



## Secret managers

FIXME document


### Bitwarden

FIXME document


### Keyring

FIXME document


### LastPass

FIXME document


### 1Password

FIXME document


### pass

FIXME document


### Vault

FIXME document


### Generic





vim: spell spelllang=en
