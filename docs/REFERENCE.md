# chezmoi reference manual

Manage your dotfiles across multiple machines, securely.

<!--- toc --->
* [Concepts](#concepts)
* [Global command line flags](#global-command-line-flags)
  * [`--color` *value*](#--color-value)
  * [`-c`, `--config` *filename*](#-c---config-filename)
  * [`--config-format` `json`|`toml`|`yaml`](#--config-format-jsontomlyaml)
  * [`-D`, `--destination` *directory*](#-d---destination-directory)
  * [`-n`, `--dry-run`](#-n---dry-run)
  * [`--force`](#--force)
  * [`-h`, `--help`](#-h---help)
  * [`-k`, `--keep-going`](#-k---keep-going)
  * [`--no-pager`](#--no-pager)
  * [`--no-tty`](#--no-tty)
  * [`-o`, `--output` *filename*](#-o---output-filename)
  * [`-R`, `--refresh-externals`](#-r---refresh-externals)
  * [`-S`, `--source` *directory*](#-s---source-directory)
  * [`--use-builtin-age` *value*](#--use-builtin-age-value)
  * [`--use-builtin-git` *value*](#--use-builtin-git-value)
  * [`-v`, `--verbose`](#-v---verbose)
  * [`--version`](#--version)
  * [`-w`, `--working-tree` *directory*](#-w---working-tree-directory)
* [Common command line flags](#common-command-line-flags)
  * [`-f`, `--format` `json`|`yaml`](#-f---format-jsonyaml)
  * [`-i`, `--include` *types*](#-i---include-types)
  * [`-r`, `--recursive`](#-r---recursive)
  * [`-x`, `--exclude` *types*](#-x---exclude-types)
* [Developer command line flags](#developer-command-line-flags)
  * [`--cpu-profile` *filename*](#--cpu-profile-filename)
  * [`--debug`](#--debug)
  * [`--gops`](#--gops)
* [Configuration file](#configuration-file)
  * [Variables](#variables)
  * [Examples](#examples)
* [Source state attributes](#source-state-attributes)
* [Target types](#target-types)
  * [Files](#files)
  * [Directories](#directories)
  * [Symbolic links](#symbolic-links)
  * [Scripts](#scripts)
  * [`symlink` mode](#symlink-mode)
* [Special files and directories](#special-files-and-directories)
  * [`.chezmoi.<format>.tmpl`](#chezmoiformattmpl)
  * [`.chezmoidata.<format>`](#chezmoidataformat)
  * [`.chezmoiexternal.<format>`](#chezmoiexternalformat)
  * [`.chezmoiignore`](#chezmoiignore)
  * [`.chezmoiremove`](#chezmoiremove)
  * [`.chezmoitemplates`](#chezmoitemplates)
  * [`.chezmoiversion`](#chezmoiversion)
* [Commands](#commands)
  * [`add` *target*...](#add-target)
  * [`apply` [*target*...]](#apply-target)
  * [`archive`](#archive)
  * [`cat` *target*...](#cat-target)
  * [`cd`](#cd)
  * [`chattr` *attributes* *target*...](#chattr-attributes-target)
  * [`completion` *shell*](#completion-shell)
  * [`data`](#data)
  * [`decrypt` [*file*...]](#decrypt-file)
  * [`diff` [*target*...]](#diff-target)
  * [`docs` [*regexp*]](#docs-regexp)
  * [`doctor`](#doctor)
  * [`dump` [*target*...]](#dump-target)
  * [`edit` [*target*...]](#edit-target)
  * [`edit-config`](#edit-config)
  * [`encrypt` [*file*...]](#encrypt-file)
  * [`execute-template` [*template*...]](#execute-template-template)
  * [`forget` *targets*](#forget-targets)
  * [`git` [*arg*...]](#git-arg)
  * [`help` [*command*...]](#help-command)
  * [`init` [*repo*]](#init-repo)
  * [`import` *filename*](#import-filename)
  * [`manage` *targets*](#manage-targets)
  * [`managed`](#managed)
  * [`merge` *target*...](#merge-target)
  * [`merge-all` [*target*...]](#merge-all-target)
  * [`purge`](#purge)
  * [`remove` *targets*](#remove-targets)
  * [`re-add`](#re-add)
  * [`rm` *targets*](#rm-targets)
  * [`secret`](#secret)
  * [`source-path` [*target*...]](#source-path-target)
  * [`state`](#state)
  * [`status`](#status)
  * [`unmanage` *target*...](#unmanage-target)
  * [`unmanaged`](#unmanaged)
  * [`update`](#update)
  * [`upgrade`](#upgrade)
  * [`verify` [*target*...]](#verify-target)
* [Editor configuration](#editor-configuration)
* [Umask configuration](#umask-configuration)
* [Template execution](#template-execution)
* [Template variables](#template-variables)
* [Template functions](#template-functions)
  * [`bitwarden` [*arg*...]](#bitwarden-arg)
  * [`bitwardenAttachment` *filename* *itemid*](#bitwardenattachment-filename-itemid)
  * [`bitwardenFields` [*arg*...]](#bitwardenfields-arg)
  * [`decrypt` *ciphertext*](#decrypt-ciphertext)
  * [`encrypt` *plaintext*](#encrypt-plaintext)
  * [`gitHubKeys` *user*](#githubkeys-user)
  * [`gopass` *gopass-name*](#gopass-gopass-name)
  * [`gopassRaw` *gopass-name*](#gopassraw-gopass-name)
  * [`include` *filename*](#include-filename)
  * [`ioreg`](#ioreg)
  * [`joinPath` *element*...](#joinpath-element)
  * [`keepassxc` *entry*](#keepassxc-entry)
  * [`keepassxcAttribute` *entry* *attribute*](#keepassxcattribute-entry-attribute)
  * [`keyring` *service* *user*](#keyring-service-user)
  * [`lastpass` *id*](#lastpass-id)
  * [`lastpassRaw` *id*](#lastpassraw-id)
  * [`lookPath` *file*](#lookpath-file)
  * [`mozillaInstallHash` *path*](#mozillainstallhash-path)
  * [`onepassword` *uuid* [*vault-uuid* [*account-name*]]](#onepassword-uuid-vault-uuid-account-name)
  * [`onepasswordDocument` *uuid* [*vault-uuid* [*account-name*]]](#onepassworddocument-uuid-vault-uuid-account-name)
  * [`onepasswordDetailsFields` *uuid* [*vault-uuid* [*account-name*]]](#onepassworddetailsfields-uuid-vault-uuid-account-name)
  * [`onepasswordItemFields` *uuid* [*vault-uuid* [*account-name*]]](#onepassworditemfields-uuid-vault-uuid-account-name)
  * [`output` *name* [*arg*...]](#output-name-arg)
  * [`pass` *pass-name*](#pass-pass-name)
  * [`passRaw` *pass-name*](#passraw-pass-name)
  * [`promptBool` *prompt* [*default*]](#promptbool-prompt-default)
  * [`promptInt` *prompt* [*default*]](#promptint-prompt-default)
  * [`promptString` *prompt* [*default*]](#promptstring-prompt-default)
  * [`secret` [*arg*...]](#secret-arg)
  * [`secretJSON` [*arg*...]](#secretjson-arg)
  * [`stat` *name*](#stat-name)
  * [`stdinIsATTY`](#stdinisatty)
  * [`vault` *key*](#vault-key)
  * [`writeToStdout` *string*...](#writetostdout-string)

---

## Concepts

chezmoi evaluates the source state for the current machine and then updates the
destination directory, where:

* The *source state* declares the desired state of your home directory,
  including templates and machine-specific configuration.

* The *source directory* is where chezmoi stores the source state, by default
  `~/.local/share/chezmoi`.

* The *target state* is the source state computed for the current machine.

* The *destination directory* is the directory that chezmoi manages, by default
  your home directory.

* A *target* is a file, directory, or symlink in the destination directory.

* The *destination state* is the current state of all the targets in the
  destination directory.

* The *config file* contains machine-specific configuration, by default it is
  `~/.config/chezmoi/chezmoi.toml`.

* The *working tree* is the git working tree. Normally it is the same as the
  source directory, but can be a parent of the source directory.

---

## Global command line flags

Command line flags override any values set in the configuration file.

### `--color` *value*

Colorize diffs, *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by `parseBool`. The default is `auto` which will colorize diffs only
if the the environment variable `$NO_COLOR` is not set and stdout is a terminal.

### `-c`, `--config` *filename*

Read the configuration from *filename*.

### `--config-format` `json`|`toml`|`yaml`

Assume the configuration file is in the given format. This is only needed if the
config filename does not have an extension, for example when it is `/dev/stdin`.

### `-D`, `--destination` *directory*

Use *directory* as the destination directory.

### `-n`, `--dry-run`

Set dry run mode. In dry run mode, the destination directory is never modified.
This is most useful in combination with the `-v` (verbose) flag to print changes
that would be made without making them.

### `--force`

Make changes without prompting.

### `-h`, `--help`

Print help.

### `-k`, `--keep-going`

Keep going as far as possible after a encountering an error.

### `--no-pager`

Do not use the pager.

### `--no-tty`

Do not attempt to get a TTY to read input and passwords. Instead, read them from
stdin.

### `-o`, `--output` *filename*

Write the output to *filename* instead of stdout.

### `-R`, `--refresh-externals`

Refresh externals cache. See `.chezmoiexternal.<format>`.

### `-S`, `--source` *directory*

Use *directory* as the source directory.

### `--use-builtin-age` *value*

Use chezmoi's builtin [age encryption](https://age-encryption.org) instead of an
external `age` command. *value* can be `on`, `off`, `auto`, or any boolean-like
value recognized by `parseBool`. The default is `auto` which will only use the
builtin age if `age.command` cannot be found in `$PATH`.

The builtin `age` command does not support passphrases, symmetric encryption, or
the use of SSH keys.

### `--use-builtin-git` *value*

Use chezmoi's builtin git instead of `git.command` for the `init` and `update`
commands. *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by `parseBool`. The default is `auto` which will only use the builtin
git if `git.command` cannot be found in `$PATH`.

### `-v`, `--verbose`

Set verbose mode. In verbose mode, chezmoi prints the changes that it is making
as approximate shell commands, and any differences in files between the target
state and the destination set are printed as unified diffs.

### `--version`

Print the version of chezmoi, the commit at which it was built, and the build
timestamp.

### `-w`, `--working-tree` *directory*

Use *directory* as the git working tree directory. By default, chezmoi searches
the source directory and then its ancestors for the first directory that
contains a `.git` directory.

---

## Common command line flags

The following flags apply to multiple commands where they are relevant.

### `-f`, `--format` `json`|`yaml`

Set the output format.

### `-i`, `--include` *types*

Only operate on target state entries of type *types*. *types* is a
comma-separated list of target states (`all`, `dirs`, `files`, `remove`,
`scripts`, `symlinks`, and `encrypted`) and can be excluded by preceding them
with a `no`. For example, `--include=dirs,files` will cause the command to apply
to directories and files only.

#### `--init`

Regenerate and reread the config file from the config file template before
computing the target state.

### `-r`, `--recursive`

Recurse into subdirectories, `true` by default.

### `-x`, `--exclude` *types*

Exclude target state entries of type *types*. *types* is a comma-separated list
of target states (`all`, `dirs`, `files`, `remove`, `scripts`, `symlinks`, and
`encrypted`). For example, `--exclude=scripts` will cause the command to not run
scripts and `--exclude=encrypted` will exclude encrypted files.

## Developer command line flags

The following flags are global but only relevant for developers and debugging.

### `--cpu-profile` *filename*

Write a [Go CPU profile](https://blog.golang.org/pprof) to *filename*.

### `--debug`

Log information helpful for debugging.

### `--gops`

Enable the [gops](https://github.com/google/gops) agent.

---

## Configuration file

chezmoi searches for its configuration file according to the [XDG Base Directory
Specification](https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html)
and supports all formats supported by
[`github.com/spf13/viper`](https://github.com/spf13/viper), namely
[JSON](https://www.json.org/json-en.html),
[TOML](https://github.com/toml-lang/toml), [YAML](https://yaml.org/), macOS
property file format, and [HCL](https://github.com/hashicorp/hcl). The basename
of the config file is `chezmoi`, and the first config file found is used.

In most installations, the config file will be read from
`~/.config/chezmoi/chezmoi.<format>`, where `<format>` is one of `json`, `toml`,
or `yaml`. The config file can be set explicitly with the `--config` command
line option. By default, the format is detected based on the extension of the
config file name, but can be overridden with the `--config-format` command line
option.

### Variables

The following configuration variables are available:

| Section        | Variable              | Type     | Default value            | Description                                            |
| -------------- | --------------------- | -------- | ------------------------ | ------------------------------------------------------ |
| Top level      | `color`               | string   | `auto`                   | Colorize output                                        |
|                | `data`                | any      | *none*                   | Template data                                          |
|                | `destDir`             | string   | `~`                      | Destination directory                                  |
|                | `encryption`          | string   | *none*                   | Encryption tool, either `age` or `gpg`                 |
|                | `format`              | string   | `json`                   | Format for data output, either `json` or `yaml`        |
|                | `mode`                | string   | `file`                   | Mode in target dir, either `file` or `symlink`         |
|                | `sourceDir`           | string   | `~/.local/share/chezmoi` | Source directory                                       |
|                | `pager`               | string   | `$PAGER`                 | Default pager                                          |
|                | `umask`               | int      | *from system*            | Umask                                                  |
|                | `useBuiltinAge`       | string   | `auto`                   | Use builtin git if `age` command is not found in $PATH |
|                | `useBuiltinGit`       | string   | `auto`                   | Use builtin git if `git` command is not found in $PATH |
|                | `workingTree`         | string   | *source directory*       | git working tree directory                             |
| `add`          | `templateSymlinks`    | bool     | `false`                  | Template symlinks to source and home dirs              |
| `age`          | `args`                | []string | *none*                   | Extra args to age CLI command                          |
|                | `command`             | string   | `age`                    | age CLI command                                        |
|                | `identity`            | string   | *none*                   | age identity file                                      |
|                | `identities`          | []string | *none*                   | age identity files                                     |
|                | `passphrase`          | bool     | `false`                  | Use age passphrase instead of identity                 |
|                | `recipient`           | string   | *none*                   | age recipient                                          |
|                | `recipients`          | []string | *none*                   | age recipients                                         |
|                | `recipientsFile`      | []string | *none*                   | age recipients file                                    |
|                | `recipientsFiles`     | []string | *none*                   | age recipients files                                   |
|                | `suffix`              | string   | `.age`                   | Suffix appended to age-encrypted files                 |
|                | `symmetric`           | bool     | `false`                  | Use age symmetric encryption                           |
| `bitwarden`    | `command`             | string   | `bw`                     | Bitwarden CLI command                                  |
| `cd`           | `args`                | []string | *none*                   | Extra args to shell in `cd` command                    |
|                | `command`             | string   | *none*                   | Shell to run in `cd` command                           |
| `diff`         | `args`                | []string | *see `diff` below*       | Extra args to external diff command                    |
|                | `command`             | string   | *none*                   | External diff command                                  |
|                | `exclude`             | []string | *none*                   | Entry types to exclude from diff                       |
|                | `pager`               | string   | *none*                   | Diff-specific pager                                    |
| `docs`         | `maxWidth`            | int      | 80                       | Maximum width of output                                |
|                | `pager`               | string   | *none*                   | Docs-specific pager                                    |
| `edit`         | `args`                | []string | *none*                   | Extra args to edit command                             |
|                | `command`             | string   | `$EDITOR` / `$VISUAL`    | Edit command                                           |
|                | `minDuration`         | duration | `1s`                     | Minimum duration for edit command                      |
| `secret`       | `command`             | string   | *none*                   | Generic secret command                                 |
| `git`          | `autoAdd `            | bool     | `false`                  | Add changes to the source state after any change       |
|                | `autoCommit`          | bool     | `false`                  | Commit changes to the source state after any change    |
|                | `autoPush`            | bool     | `false`                  | Push changes to the source state after any change      |
|                | `command`             | string   | `git`                    | Source version control system                          |
| `gopass`       | `command`             | string   | `gopass`                 | gopass CLI command                                     |
| `gpg`          | `args`                | []string | *none*                   | Extra args to GPG CLI command                          |
|                | `command`             | string   | `gpg`                    | GPG CLI command                                        |
|                | `recipient`           | string   | *none*                   | GPG recipient                                          |
|                | `suffix`              | string   | `.asc`                   | Suffix appended to GPG-encrypted files                 |
|                | `symmetric`           | bool     | `false`                  | Use symmetric GPG encryption                           |
| `interpreters` | *extension*`.args`    | []string | *none*                   | See section on "Scripts on Windows"                    |
|                | *extension*`.command` | string   | *special*                | See section on "Scripts on Windows"                    |
| `keepassxc`    | `args`                | []string | *none*                   | Extra args to KeePassXC CLI command                    |
|                | `command`             | string   | `keepassxc-cli`          | KeePassXC CLI command                                  |
|                | `database`            | string   | *none*                   | KeePassXC database                                     |
| `lastpass`     | `command`             | string   | `lpass`                  | Lastpass CLI command                                   |
| `merge`        | `args`                | []string | *see `merge` below*      | Args to 3-way merge command                            |
|                | `command`             | string   | `vimdiff`                | 3-way merge command                                    |
| `onepassword`  | `cache`               | bool     | `true`                   | Enable optional caching provided by `op`               |
|                | `command`             | string   | `op`                     | 1Password CLI command                                  |
| `pass`         | `command`             | string   | `pass`                   | Pass CLI command                                       |
| `template`     | `options`             | []string | `["missingkey=error"]`   | Template options                                       |
| `vault`        | `command`             | string   | `vault`                  | Vault CLI command                                      |

### Examples

#### JSON

```json
{
    "sourceDir": "/home/user/.dotfiles",
    "git": {
        "autoPush": true
    }
}
```

#### TOML

```toml
sourceDir = "/home/user/.dotfiles"
[git]
    autoPush = true
```

#### YAML

```yaml
sourceDir: /home/user/.dotfiles
git:
    autoPush: true
```

---

## Source state attributes

chezmoi stores the source state of files, symbolic links, and directories in
regular files and directories in the source directory (`~/.local/share/chezmoi`
by default). This location can be overridden with the `-S` flag or by giving a
value for `sourceDir` in `~/.config/chezmoi/chezmoi.toml`. Directory targets are
represented as directories in the source state. All other target types are
represented as files in the source state. Some state is encoded in the source
names.

The following prefixes and suffixes are special, and are collectively referred
to as "attributes":

| Prefix       | Effect                                                                          |
| ------------ | ------------------------------------------------------------------------------- |
| `after_`     | Run script after updating the destination.                                      |
| `before_`    | Run script before updating the destination.                                     |
| `create_`    | Ensure that the file exists, and create it with contents if it does not.        |
| `dot_`       | Rename to use a leading dot, e.g. `dot_foo` becomes `.foo`.                     |
| `empty_`     | Ensure the file exists, even if is empty. By default, empty files are removed.  |
| `encrypted_` | Encrypt the file in the source state.                                           |
| `exact_`     | Remove anything not managed by chezmoi.                                         |
| `executable_`| Add executable permissions to the target file.                                  |
| `literal_`   | Stop parsing prefix attributes.                                                 |
| `modify_`    | Treat the contents as a script that modifies an existing file.                  |
| `once_`      | Only run the script if it has not been run before.                              |
| `onchange_`  | Only run the script if its contents have changed from the last time it was run. |
| `private_`   | Remove all group and world permissions from the target file or directory.       |
| `readonly_`  | Remove all write permissions from the target file or directory.                 |
| `remove_`    | Remove the entry if it exists.                                                  |
| `run_`       | Treat the contents as a script to run.                                          |
| `symlink_`   | Create a symlink instead of a regular file.                                     |

| Suffix     | Effect                                               |
| ---------- | ---------------------------------------------------- |
| `.literal` | Stop parsing suffix attributes.                      |
| `.tmpl`    | Treat the contents of the source file as a template. |

Different target types allow different prefixes and suffixes. The order of
prefixes is important.

| Target type   | Source type | Allowed prefixes in order                                               | Allowed suffixes |
| ------------- | ----------- | ----------------------------------------------------------------------- | ---------------- |
| Directory     | Directory   | `exact_`, `private_`, `readonly_`, `dot_`                               | *none*           |
| Regular file  | File        | `encrypted_`, `private_`, `executable_`, `dot_`                         | `.tmpl`          |
| Create file   | File        | `create_`, `encrypted_`, `private_`, `readonly_`, `executable_`, `dot_` | `.tmpl`          |
| Modify file   | File        | `modify_`, `encrypted_`, `private_`, `readonly_`, `executable_`, `dot_` | `.tmpl`          |
| Remove        | File        | `remove_`, `dot_`                                                       | *none*           |
| Script        | File        | `run_`, `once_` or `onchange_`, `before_` or `after_`                   | `.tmpl`          |
| Symbolic link | File        | `symlink_`, `dot_`,                                                     | `.tmpl`          |

The `literal_` prefix and `.literal` suffix can appear anywhere and stop
attribute parsing. This permits filenames that would otherwise conflict with
chezmoi's attributes to be represented.

In addition, if the source file is encrypted, the suffix `.age` (when age
encryption is used) or `.asc` (when gpg encryption is used) is stripped. These
suffixes can be overridden with the `age.suffix` and `gpg.suffix` configuration
variables.

chezmoi ignores all files and directories in the source directory that begin
with a `.` with the exception of files and directories that begin with
`.chezmoi`.

---

## Target types

chezmoi will create, update, and delete files, directories, and symbolic links
in the destination directory, and run scripts. chezmoi deterministically
performs actions in ASCII order of their target name. For example, given a file
`dot_a`, a script `run_z`, and a directory `exact_dot_c`, chezmoi will first
create `.a`, create `.c`, and then execute `run_z`.

---

### Files

Files are represented by regular files in the source state. The `encrypted_`
attribute determines whether the file in the source state is encrypted. The
`executable_` attribute will set the executable bits in the the target state,
and the `private_` attribute will clear all group and world permissions. The
`readonly_` attribute will clear all write permission bits in the target state.
Files with the `.tmpl` suffix will be interpreted as templates. If the target
contents are empty then the file will be removed, unless it has an `empty_`
prefix.

#### Create file

Files with the `create_` prefix will be created in the target state with the
contents of the file in the source state if they do not already exist. If the
file in the destination state already exists then its contents will be left
unchanged.

#### Modify file

Files with the `modify_` prefix are treated as scripts that modify an existing
file. The contents of the existing file (which maybe empty if the existing file
does not exist or is empty) are passed to the script's standard input, and the
new contents are read from the scripts standard output.

---

#### Remove entry

Files with the `remove_` prefix will cause the corresponding entry (file,
directory, or symlink) to be removed in the target state.

---

### Directories

Directories are represented by regular directories in the source state. The
`exact_` attribute causes chezmoi to remove any entries in the target state that
are not explicitly specified in the source state, and the `private_` attribute
causes chezmoi to clear all group and world permissions. The `readonly_`
attribute will clear all write permission bits.

---

### Symbolic links

Symbolic links are represented by regular files in the source state with the
prefix `symlink_`. The contents of the file will have a trailing newline
stripped, and the result be interpreted as the target of the symbolic link.
Symbolic links with the `.tmpl` suffix in the source state are interpreted as
templates. If the target of the symbolic link is empty or consists only of
whitespace, then the target is removed.

---

### Scripts

Scripts are represented as regular files in the source state with prefix `run_`.
The file's contents (after being interpreted as a template if it has a `.tmpl`
suffix) are executed.

Scripts are executed on every `chezmoi apply`, unless they have the `once_` or
`onchange_` attribute. `run_once_` scripts are only executed if a script with
the same contents has not been run before, for example if its contents has
changed. `run_onchange_` scripts are executed whenever their contents change,
even if a script with the same contents has run before.

Scripts with the `before_` attribute are executed before any files, directories,
or symlinks are updated. Scripts with the `after_` attribute are executed after
all files, directories, and symlinks have been updated. Scripts without an
`before_` or `after_` attribute are executed in ASCII order of their target
names with respect to files, directories, and symlinks.

Scripts will normally run with their working directory set to their equivalent
location in the destination directory. For example, a script in
`~/.local/share/chezmoi/dir/run_script` will be run with a working directory of
`~/dir`. If the equivalent location in the destination directory either does not
exist or is not a directory, then chezmoi will walk up the script's directory
hierarchy and run the script in the first directory that exists and is a
directory.

#### Scripts on Windows

<!-- FIXME: some of the following needs to be moved to the how-to -->

The execution of scripts on Windows depends on the script's file extension.
Windows will natively execute scripts with a `.bat`, `.cmd`, `.com`, and `.exe`
extensions. Other extensions require an interpreter, which must be in your
`%PATH%`.

The default script interpreters are:

| Extension | Command      | Arguments |
| --------- | ------------ | --------- |
| `.pl`     | `perl`       | *none*    |
| `.py`     | `python`     | *none*    |
| `.ps1`    | `powershell` | `-NoLogo` |
| `.rb`     | `ruby`       | *none*    |

Script interpreters can be added or overridden with the
`interpreters.`*extension* section in the configuration file. Note that the
leading `.` is dropped from *extension*.

For example to change the Python interpreter to `C:\Python39\python.exe` and add
a Tcl/Tk interpreter, include the following in `~/.config/chezmoi/chezmoi.toml`:

```toml
[interpreters.py]
    command = 'C:\Python39\python.exe'
[interpreters.tcl]
    command = "tclsh"
```

If the script in the source state is a template (with a `.tmpl` extension), then
chezmoi will strip the `.tmpl` extension and use the next remaining extension to
determine the interpreter to use.

---

### `symlink` mode

By default, chezmoi will create regular files and directories. Setting `mode =
"symlink"` will make chezmoi behave more like a dotfile manager that uses
symlinks by default, i.e. `chezmoi apply` will make dotfiles symlinks to files
in the source directory if the target is a regular file and is not
encrypted, executable, private, or a template.

---

## Special files and directories

All files and directories in the source state whose name begins with `.` are
ignored by default, unless they are one of the special files listed here.

---

### `.chezmoi.<format>.tmpl`

If a file called `.chezmoi.<format>.tmpl` exists then `chezmoi init` will use it
to create an initial config file. *format* must be one of the the supported
config file formats.

#### `.chezmoi.<format>.tmpl` examples

```
{{ $email := promptString "email" -}}
data:
    email: {{ $email | quote }}
```

---

### `.chezmoidata.<format>`

If a file called `.chezmoidata.<format>` exists in the source state, it is interpreted
as a datasource available in most [templates](TEMPLATING.md#template-data).

#### `.chezmoidata.<format>` examples

If `.chezmoidata.toml` contains the following (and no variable is overwritten in later stages):

```toml
editor = "nvim"
[directions]
	up = "k"
	down = "j"
	right = "l"
	left = "h"
```

Then the following template:

```
EDITOR={{ .editor }}
MOVE_UP={{ .directions.up }}
MOVE_DOWN={{ .directions.down }}
MOVE_RIGHT={{ .directions.right }}
MOVE_LEFT={{ .directions.left }}
```

Will result in:

```
EDITOR=nvim
MOVE_UP=k
MOVE_DOWN=j
MOVE_RIGHT=l
MOVE_LEFT=h
```

---

### `.chezmoiexternal.<format>`

If a file called `.chezmoiexternal.<format>` exists in the source state, it is
interpreted as a list of external files and archives to be included as if they
were in the source state.

`.chezmoiexternal.<format>` is interpreted as a template. This allows different
externals to be included on different machines.

Entries are indexed by target name relative to the directory of the
`.chezmoiexternal.<format>` file, and must have a `type` and a `url` field.
`type` can be either `file` or `archive`. All of the entries parent directories
must be defined in the source state. chezmoi will not create parent directories
automatically.

Entries may have the following fields:

| Variable          | Type     | Default value | Description                                                   |
| ----------------- | -------- | ------------- | ------------------------------------------------------------- |
| `type`            | string   | *none*        | External type (`file` or `archive`)                           |
| `encrypted`       | bool     | `false`       | Whether the external is encrypted                             |
| `exact`           | bool     | `false`       | Add `exact_` attribute to directories in archive              |
| `executable`      | bool     | `false`       | Add `executable_` attribute to file                           |
| `filter.command`  | string   | *none*        | Command to filter contents                                    |
| `filter.args`     | []string | *none*        | Extra args to command to filter contents                      |
| `format`          | string   | *autodetect*  | Format of archive                                             |
| `refreshPeriod`   | duration | `0`           | Refresh period                                                |
| `stripComponents` | int      | `0`           | Number of leading directory components to strip from archives |
| `url`             | string   | *none*        | URL                                                           |

The optional boolean `encrypted` field specifies whether the file or archive
is encrypted.

If optional string `filter.command` and array of strings `filter.args` are
specified, the the file or archive is filtered by piping it into the command's
standard input and reading the command's standard output.

If `type` is `file` then the target is a file with the contents of `url`. The
optional boolean field `executable` may be set, in which case the target file
will be executable.

If `type` is `archive` then the target is a directory with the contents of the
archive at `url`. The optional boolean field `exact` may be set, in which case
the directory and all subdirectories will be treated as exact directories, i.e.
`chezmoi apply` will remove entries not present in the archive. The optional
integer field `stripComponents` will remove leading path components from the
members of archive. The optional string field `format` sets the archive format.
The supported archive formats are `tar`, `tar.gz`, `tgz`, `tar.bz2`, `tbz2`, and
`zip`. If `format` is not specified then chezmoi will guess the format using
firstly the path of the URL and secondly its contents.

By default, chezmoi will cache downloaded URLs. The optional duration
`refreshPeriod` field specifies how often chezmoi will re-download the URL. The
default is zero meaning that chezmoi will never re-download unless forced. To
force chezmoi to re-download URLs, pass the `-R`/`--refresh-externals` flag.
Suitable refresh periods include one day (`24h`), one week (`168h`), or four
weeks (`672h`).

#### `.chezmoiexternal.<format>` examples

```toml
[".vim/autoload/plug.vim"]
    type = "file"
    url = "https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim"
    refreshPeriod = "168h"
[".oh-my-zsh"]
    type = "archive"
    url = "https://github.com/ohmyzsh/ohmyzsh/archive/master.tar.gz"
    exact = true
    stripComponents = 1
    refreshPeriod = "168h"
[".oh-my-zsh/custom/plugins/zsh-syntax-highlighting"]
    type = "archive"
    url = "https://github.com/zsh-users/zsh-syntax-highlighting/archive/master.tar.gz"
    exact = true
    stripComponents = 1
    refreshPeriod = "168h"
[".oh-my-zsh/custom/themes/powerlevel10k"]
    type = "archive"
    url = "https://github.com/romkatv/powerlevel10k/archive/v1.15.0.tar.gz"
    exact = true
    stripComponents = 1
```

---

### `.chezmoiignore`

If a file called `.chezmoiignore` exists in the source state then it is
interpreted as a set of patterns to ignore. Patterns are matched using
[`doublestar.Match`](https://pkg.go.dev/github.com/bmatcuk/doublestar/v4#Match)
and match against the target path, not the source path.

Patterns can be excluded by prefixing them with a `!` character. All excludes
take priority over all includes.

Comments are introduced with the `#` character and run until the end of the
line.

`.chezmoiignore` is interpreted as a template. This allows different files to be
ignored on different machines.

`.chezmoiignore` files in subdirectories apply only to that subdirectory.

#### `.chezmoiignore` examples

```
README.md

*.txt   # ignore *.txt in the target directory
*/*.txt # ignore *.txt in subdirectories of the target directory
        # but not in subdirectories of subdirectories;
        # so a/b/c.txt would *not* be ignored

backups/   # ignore backups folder in chezmoi directory and all its contents
backups/** # ignore all contents of backups folder in chezmoi directory
           # but not backups folder itself

{{- if ne .email "firstname.lastname@company.com" }}
# Ignore .company-directory unless configured with a company email
.company-directory # note that the pattern is not dot_company-directory
{{- end }}

{{- if ne .email "me@home.org }}
.personal-file
{{- end }}
```

---

### `.chezmoiremove`

If a file called `.chezmoiremove` exists in the source state then it is
interpreted as a list of targets to remove. `.chezmoiremove` is interpreted as a
template.

---

### `.chezmoitemplates`

If a directory called `.chezmoitemplates` exists, then all files in this
directory are parsed as templates are available as templates with a name equal
to the relative path to the `.chezmoitemplates` directory.

The [`template` action](https://pkg.go.dev/text/template#hdr-Actions) can be
used to include these templates in another template. The value of `.` must be
set explicitly if needed, otherwise the template will be executed with `nil`
data.

#### `.chezmoitemplates` examples

Given:

`.chezmoitemplates/foo`:
```
{{ if true }}bar{{ end }}
```

`dot_config.tmpl`:
```
{{ template "foo" . }}
```

The target state of `.config` will be `bar`.

---

### `.chezmoiversion`

If a file called `.chezmoiversion` exists, then its contents are interpreted as
a semantic version defining the minimum version of chezmoi required to interpret
the source state correctly. chezmoi will refuse to interpret the source state if
the current version is too old.

#### `.chezmoiversion` examples

```
1.5.0
```

---

## Commands

---

### `add` *target*...

Add *target*s to the source state. If any target is already in the source state,
then its source state is replaced with its current state in the destination
directory.

#### `--autotemplate`

Automatically generate a template by replacing strings with variable names from
the `data` section of the config file. Longer substitutions occur before shorter
ones. This implies the `--template` option.

#### `-e`, `--empty`

Set the `empty` attribute on added files.

#### `-f`, `--force`

Add *targets*, even if doing so would cause a source template to be overwritten.

#### `--follow`

If the last part of a target is a symlink, add the target of the symlink instead
of the symlink itself.

#### `--exact`

Set the `exact` attribute on added directories.

#### `-i`, `--include` *types*

Only add entries of type *types*.

#### `-p`, `--prompt`

Interactively prompt before adding each file.

#### `-r`, `--recursive`

Recursively add all files, directories, and symlinks.

#### `-T`, `--template`

Set the `template` attribute on added files and symlinks.

#### `--template-symlinks`

When adding symlink to an absolute path in the source directory or destination
directory, create a symlink template with `.chezmoi.sourceDir` or
`.chezmoi.homeDir`. This is useful for creating portable absolute symlinks.

#### `add` examples

```console
$ chezmoi add ~/.bashrc
$ chezmoi add ~/.gitconfig --template
$ chezmoi add ~/.vim --recursive
$ chezmoi add ~/.oh-my-zsh --exact --recursive
```

---

### `apply` [*target*...]

Ensure that *target*... are in the target state, updating them if necessary. If
no targets are specified, the state of all targets are ensured. If a target has
been modified since chezmoi last wrote it then the user will be prompted if they
want to overwrite the file.

#### `-i`, `--include` *types*

Only add entries of type *types*.

#### `--source-path`

Specify targets by source path, rather than target path. This is useful for
applying changes after editing.

#### `apply` examples

```console
$ chezmoi apply
$ chezmoi apply --dry-run --verbose
$ chezmoi apply ~/.bashrc
```

---

### `archive`

Generate an archive of the target state. This can be piped into `tar` to inspect
the target state.

#### `-f`, `--format` `tar`|`tar.gz`|`tgz`|`zip`

Write the archive in *format*. If `--output` is set the format is guessed from
the extension, otherwise the default is `tar`.

#### `-i`, `--include` *types*

Only include entries of type *types*.

#### `-z`, `--gzip`

Compress the archive with gzip. This is automatically set if the format is
`tar.gz` or `tgz` and is ignored if the format is `zip`.

#### `archive` examples

```console
$ chezmoi archive | tar tvf -
$ chezmoi archive --output=dotfiles.tar.gz
$ chezmoi archive --output=dotfiles.zip
```

---

### `cat` *target*...

Write the target contents of *target*s to stdout. *targets* must be files,
scripts, or symlinks. For files, the target file contents are written. For
scripts, the script's contents are written. For symlinks, the target target is
written.

#### `cat` examples

```console
$ chezmoi cat ~/.bashrc
```

---

### `cd`

Launch a shell in the working tree (typically the source directory). chezmoi
will launch the command set by the `cd.command` configuration variable with any
extra arguments specified by `cd.args`. If this is not set, chezmoi will attempt
to detect your shell and will finally fall back to an OS-specific default.

#### `cd` examples

```console
$ chezmoi cd
```

---

### `chattr` *attributes* *target*...

Change the attributes of *target*s. *attributes* specifies which attributes to
modify. Add attributes by specifying them or their abbreviations directly,
optionally prefixed with a plus sign (`+`). Remove attributes by prefixing them
or their attributes with the string `no` or a minus sign (`-`). The available
attributes and their abbreviations are:

| Attribute    | Abbreviation |
| ------------ | ------------ |
| `after`      | `a`          |
| `before`     | `b`          |
| `empty`      | `e`          |
| `encrypted`  | *none*       |
| `exact`      | *none*       |
| `executable` | `x`          |
| `once`       | `o`          |
| `private`    | `p`          |
| `readonly`   | `r`          |
| `template`   | `t`          |

Multiple attributes modifications may be specified by separating them with a
comma (`,`). If you use the `-`*attr* form then you must put *attributes* after
a `--` to prevent chezmoi from interpreting `-`*attr as an option.

#### `chattr` examples

```console
$ chezmoi chattr template ~/.bashrc
$ chezmoi chattr noempty ~/.profile
$ chezmoi chattr private,template ~/.netrc
$ chezmoi chattr -- -x ~/.zshrc
```

---

### `completion` *shell*

Generate shell completion code for the specified shell (`bash`, `fish`,
`powershell`, or `zsh`).

#### `completion` examples

```console
$ chezmoi completion bash
$ chezmoi completion fish --output=~/.config/fish/completions/chezmoi.fish
```

---

### `data`

Write the computed template data to stdout.

#### `-f`, `--format` `json`|`yaml`

Set the output format.

#### `data` examples

```console
$ chezmoi data
$ chezmoi data --format=yaml
```

---

### `decrypt` [*file*...]

Decrypt *file*s using chezmoi's configured encryption. If no files are given,
decrypt the standard input. The decrypted result is written to the standard
output or a file if the `--output` flag is set.

---

### `diff` [*target*...]

Print the difference between the target state and the destination state for
*target*s. If no targets are specified, print the differences for all targets.

If a `diff.pager` command is set in the configuration file then the output will
be piped into it.

If `diff.command` is set then it will be invoked to show individual file
differences with `diff.args` passed as arguments. Each element of `diff.args` is
interpreted as a template with the variables `.Destination` and `.Target`
available corresponding to the path of the file in the source and target state
respectively. The default value of `diff.args` is `["{{ .Destination }}", "{{
.Target }}"]`. If `diff.args` does not contain any template arguments then `{{
.Destination }}` and `{{ .Target }}` will be appended automatically.

#### `--pager` *pager*

Pager to use for output.

#### `--use-builtin-diff`

Use chezmoi's builtin diff, even if the `diff.command` configuration variable is
set.

#### `diff` examples

```console
$ chezmoi diff
$ chezmoi diff ~/.bashrc
```

---

### `docs` [*regexp*]

Print the documentation page matching the regular expression *regexp*. Matching
is case insensitive. If no pattern is given, print `REFERENCE.md`.

#### `--pager` *pager*

Pager to use for output.

#### `docs` examples

```console
$ chezmoi docs
$ chezmoi docs faq
$ chezmoi docs howto
```

---

### `doctor`

Check for potential problems.

#### `doctor` examples

```console
$ chezmoi doctor
```

---

### `dump` [*target*...]

Dump the target state of *target*s. If no targets are specified, then the entire
target state.

#### `-f`, `--format` `json`|`yaml`

Set the output format.

#### `-i`, `--include` *types*

Only include entries of type *types*.

#### `dump` examples

```console
$ chezmoi dump ~/.bashrc
$ chezmoi dump --format=yaml
```

---

### `edit` [*target*...]

Edit the source state of *target*s, which must be files or symlinks. If no
targets are given then the working tree of the source directory is opened.

The editor used is the first non-empty string of the `edit.command`
configuration variable, the `$VISUAL` environment variable, the `$EDITOR`
environment variable. If none are set then chezmoi falls back to `notepad.exe`
on Windows systems and `vi` on non-Windows systems.

When the `edit.command` configuration variable is used, extra arguments can be
passed to the editor with the `editor.args` configuration variable.

Encrypted files are decrypted to a private temporary directory and the editor is
invoked with the decrypted file. When the editor exits the edited decrypted file
is re-encrypted and replaces the original file in the source state.

If the operating system supports hard links, then the edit command invokes the
editor with filenames which match the target filename.

chezmoi will emit a warning if the editor returns in less than
`edit.minDuration` (default `1s`). To disable this warning, set
`edit.minDuration` to `0`.

#### `-a`, `--apply`

Apply target immediately after editing. Ignored if there are no targets.

#### `edit` examples

```console
$ chezmoi edit ~/.bashrc
$ chezmoi edit ~/.bashrc --apply
$ chezmoi edit
```

---

### `edit-config`

Edit the configuration file.

#### `edit-config` examples

```console
$ chezmoi edit-config
```

---

### `encrypt` [*file*...]

Encrypt *file*s using chezmoi's configured encryption. If no files are given,
encrypt the standard input. The encrypted result is written to the standard
output or a file if the `--output` flag is set.

---

### `execute-template` [*template*...]

Execute *template*s. This is useful for testing templates or for calling chezmoi
from other scripts. *templates* are interpreted as literal templates, with no
whitespace added to the output between arguments. If no templates are specified,
the template is read from stdin.

#### `--init`, `-i`

Include simulated functions only available during `chezmoi init`.

#### `--promptBool` *pairs*

Simulate the `promptBool` function with a function that returns values from
*pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptBool` is called with a *prompt* that does not match any of *pairs*, then
it returns false.

#### `--promptInt` *pairs*

Simulate the `promptInt` function with a function that returns values from
*pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptInt` is called with a *prompt* that does not match any of *pairs*, then
it returns zero.

#### `--promptString`, `-p` *pairs*

Simulate the `promptString` function with a function that returns values from
*pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptString` is called with a *prompt* that does not match any of *pairs*,
then it returns *prompt* unchanged.

#### `--stdinisatty` *bool*

Simulate the `stdinIsATTY` function by returning *bool*.

#### `execute-template` examples

```console
$ chezmoi execute-template '{{ .chezmoi.sourceDir }}'
$ chezmoi execute-template '{{ .chezmoi.os }}' / '{{ .chezmoi.arch }}'
$ echo '{{ .chezmoi | toJson }}' | chezmoi execute-template
$ chezmoi execute-template --init --promptString email=me@home.org < ~/.local/share/chezmoi/.chezmoi.toml.tmpl
```

---

### `forget` *targets*

Remove *targets* from the source state, i.e. stop managing them.

#### `forget` examples

```console
$ chezmoi forget ~/.bashrc
```

---

### `git` [*arg*...]

Run `git` *arg*s in the working tree (typically the source directory). Note that
flags in *arguments* must occur after `--` to prevent chezmoi from interpreting
them.

#### `git` examples

```console
$ chezmoi git add .
$ chezmoi git add dot_gitconfig
$ chezmoi git -- commit -m "Add .gitconfig"
```

---

### `help` [*command*...]

Print the help associated with *command*, or general help if no command is
given.

---

### `init` [*repo*]

Setup the source directory, generate the config file, and optionally update the
destination directory to match the target state. *repo* is expanded to a full
git repo URL, using HTTPS by default, or SSH if the `--ssh` option is specified,
according to the following patterns:

| Pattern            | HTTPS Repo                             | SSH repo                           |
| ------------------ | -------------------------------------- | ---------------------------------- |
| `user`             | `https://github.com/user/dotfiles.git` | `git@github.com:user/dotfiles.git` |
| `user/repo`        | `https://github.com/user/repo.git`     | `git@github.com:user/repo.git`     |
| `site/user/repo`   | `https://site/user/repo.git`           | `git@site:user/repo.git`           |
| `~sr.ht/user`      | `https://git.sr.ht/~user/dotfiles`     | `git@git.sr.ht:~user/dotfiles.git` |
| `~sr.ht/user/repo` | `https://git.sr.ht/~user/repo`         | `git@git.sr.ht:~/user/repo.git`    |

First, if the source directory is not already contain a repository, then if
*repo* is given it is checked out into the source directory, otherwise a new
repository is initialized in the source directory.

Second, if a file called `.chezmoi.<format>.tmpl` exists, where `<format>` is
one of the supported file formats (e.g. `json`, `toml`, or `yaml`) then a new
configuration file is created using that file as a template.

Then, if the `--apply` flag is passed, `chezmoi apply` is run.

Then, if the `--purge` flag is passed, chezmoi will remove the source directory
and its config directory.

Finally, if the `--purge-binary` is passed, chezmoi will attempt to remove its
own binary.

#### `--apply`

Run `chezmoi apply` after checking out the repo and creating the config file.

#### `--branch` *branch*

Check out *branch* instead of the default branch.

#### `--config-path` *path*

Write the generated config file to *path* instead of the default location.

#### `--data` *bool*

Include existing template data when creating the config file. This defaults to
`true`. Set this to `false` to simulate creating the config file with no
existing template data.

#### `--depth` *depth*

Clone the repo with depth *depth*.

#### `--one-shot`

`--one-shot` is the equivalent of `--apply`, `--depth=1`, `--force`, `--purge`,
and `--purge-binary`. It attempts to install your dotfiles with chezmoi and then
remove all traces of chezmoi from the system. This is useful for setting up
temporary environments (e.g. Docker containers).

#### `--purge`

Remove the source and config directories after applying.

#### `--purge-binary`

Attempt to remove the chezmoi binary after applying.

#### `--ssh`

Guess an SSH repo URL instead of an HTTPS repo.

#### `init` examples

```console
$ chezmoi init user
$ chezmoi init user --apply
$ chezmoi init user --apply --purge
$ chezmoi init user/dots
$ chezmoi init gitlab.com/user
```

---

### `import` *filename*

Import the source state from an archive file in to a directory in the source
state. This is primarily used to make subdirectories of your home directory
exactly match the contents of a downloaded archive. You will generally always
want to set the `--destination`, `--exact`, and `--remove-destination` flags.

The supported archive formats are `tar`, `tar.gz`, `tgz`, `tar.bz2`, `tbz2`, and
`zip`.

#### `--destination` *directory*

Set the destination (in the source state) where the archive will be imported.

#### `--exact`

Set the `exact` attribute on all imported directories.

#### `-r`, `--remove-destination`

Remove destination (in the source state) before importing.

#### `--strip-components` *n*

Strip *n* leading components from paths.

#### `import` examples

```console
$ curl -s -L -o ${TMPDIR}/oh-my-zsh-master.tar.gz https://github.com/ohmyzsh/ohmyzsh/archive/master.tar.gz
$ mkdir -p $(chezmoi source-path)/dot_oh-my-zsh
$ chezmoi import --strip-components 1 --destination ~/.oh-my-zsh ${TMPDIR}/oh-my-zsh-master.tar.gz
```

---

### `manage` *targets*

`manage` is an alias for `add` for symmetry with `unmanage`.

---

### `managed`

List all managed entries in the destination directory in alphabetical order.

#### `-i`, `--include` *types*

Only include entries of type *types*.

#### `managed` examples

```console
$ chezmoi managed
$ chezmoi managed --include=files
$ chezmoi managed --include=files,symlinks
$ chezmoi managed -i d
$ chezmoi managed -i d,f
```

---

### `merge` *target*...

Perform a three-way merge between the destination state, the target state, and
the source state for each *target*. The merge tool is defined by the
`merge.command` configuration variable, and defaults to `vimdiff`. If multiple
targets are specified the merge tool is invoked separately and sequentially for
each target. If the target state cannot be computed (for example if source is a
template containing errors or an encrypted file that cannot be decrypted) a
two-way merge is performed instead.

The order of arguments to `merge.command` is set by `merge.args`. Each argument
is interpreted as a template with the variables `.Destination`, `.Source`, and
`.Target` available corresponding to the path of the file in the destination
state, the source state, and the target state respectively. The default value of
`merge.args` is `["{{ .Destination }}", "{{ .Source }}", "{{ .Target }}"]`. If
`merge.args` does not contain any template arguments then `{{ .Destination }}`,
`{{ .Source }}`, and `{{ .Target }}` will be appended automatically.

#### `merge` examples

```console
$ chezmoi merge ~/.bashrc
```

---

### `merge-all` [*target*...]

Perform a three-way merge for file whose actual state does not match its target
state. The merge is performed with `chezmoi merge`.

#### `merge-all` examples

```console
$ chezmoi merge-all
```

---

### `purge`

Remove chezmoi's configuration, state, and source directory, but leave the
target state intact.

#### `-f`, `--force`

Remove without prompting.

#### `purge` examples

```console
$ chezmoi purge
$ chezmoi purge --force
```

---

### `remove` *targets*

Remove *targets* from both the source state and the destination directory.

#### `-f`, `--force`

Remove without prompting.

---

### `re-add`

Re-add all modified files in the target state. chezmoi will not overwrite
templates, and all entries that are not files are ignored.

#### `re-add` examples

```console
$ chezmoi re-add
```

### `rm` *targets*

`rm` is an alias for `remove`.

---

### `secret`

Run a secret manager's CLI, passing any extra arguments to the secret manager's
CLI. This is primarily for verifying chezmoi's integration with your secret
manager. Normally you would use template functions to retrieve secrets. Note
that if you want to pass flags to the secret manager's CLI you will need to
separate them with `--` to prevent chezmoi from interpreting them.

To get a full list of available commands run:

```console
$ chezmoi secret help
```

#### `secret` examples

```console
$ chezmoi secret keyring set --service=service --user=user --value=password
$ chezmoi secret keyring get --service=service --user=user
```

---

### `source-path` [*target*...]

Print the path to each target's source state. If no targets are specified then
print the source directory.

#### `source-path` examples

```console
$ chezmoi source-path
$ chezmoi source-path ~/.bashrc
```

---

### `state`

Manipulate the persistent state.

#### `state` examples

```console
$ chezmoi state data
$ chezmoi state delete --bucket=bucket --key=key
$ chezmoi state dump
$ chezmoi state get --bucket=bucket --key=key
$ chezmoi state set --bucket=bucket --key=key --value=value
$ chezmoi state reset
```

---

### `status`

Print the status of the files and scripts managed by chezmoi in a format similar
to [`git status`](https://git-scm.com/docs/git-status).

The first column of output indicates the difference between the last state
written by chezmoi and the actual state. The second column indicates the
difference between the actual state and the target state.

#### `-i`, `--include` *types*

Only include entries of type *types*.

#### `status` examples

```console
$ chezmoi status
```

---

### `unmanage` *target*...

`unmanage` is an alias for `forget` for symmetry with `manage`.

### `unmanaged`

List all unmanaged files in the destination directory.

#### `unmanaged` examples

```console
$ chezmoi unmanaged
```

---

### `update`

Pull changes from the source repo and apply any changes.

#### `-i`, `--include` *types*

Only update entries of type *types*.

#### `update` examples

```console
$ chezmoi update
```

---

### `upgrade`

Upgrade chezmoi by downloading and installing the latest released version. This
will call the GitHub API to determine if there is a new version of chezmoi
available, and if so, download and attempt to install it in the same way as
chezmoi was previously installed.

If the any of the `$CHEZMOI_GITHUB_ACCESS_TOKEN`, `$GITHUB_ACCESS_TOKEN`, or
`$GITHUB_TOKEN` environment variables are set, then the first value found will
be used to authenticate requests to the GitHub API, otherwise unauthenticated
requests are used which are subject to stricter [rate
limiting](https://developer.github.com/v3/#rate-limiting). Unauthenticated
requests should be sufficient for most cases.

---

### `verify` [*target*...]

Verify that all *target*s match their target state. chezmoi exits with code 0
(success) if all targets match their target state, or 1 (failure) otherwise. If
no targets are specified then all targets are checked.

#### `-i`, `--include` *types*

Only include entries of type *types*.

#### `verify` examples

```console
$ chezmoi verify
$ chezmoi verify ~/.bashrc
```

---

## Editor configuration

The `edit` and `edit-config` commands use the editor specified by the `VISUAL`
environment variable, the `$EDITOR` environment variable, or `vi`, whichever is
specified first.

---

## Umask configuration

By default, chezmoi uses your current umask as set by your operating system and
shell. chezmoi only stores crude permissions in its source state, namely in the
`executable`  and `private` attributes, corresponding to the umasks of `0o111`
and `0o077` respectively.

For machine-specific control of umask, set the `umask` configuration variable in
chezmoi's configuration file, for example:

```toml
umask = 0o22
```

---

## Template execution

chezmoi executes templates using
[`text/template`](https://pkg.go.dev/text/template). The result is treated
differently depending on whether the target is a file or a symlink.

If target is a file, then:

* If the result is an empty string, then the file is removed.
* Otherwise, the target file contents are result.

If the target is a symlink, then:

* Leading and trailing whitespace are stripped from the result.
* If the result is an empty string, then the symlink is removed.
* Otherwise, the target symlink target is the result.

chezmoi executes templates using `text/template`'s `missingkey=error` option,
which means that misspelled or missing keys will raise an error. This can be
overridden by setting a list of options in the configuration file, for example:

```toml
[template]
    options = ["missingkey=zero"]
```

For a full list of options, see
[`Template.Option`](https://pkg.go.dev/text/template?tab=doc#Template.Option).

---

## Template variables

chezmoi provides the following automatically-populated variables:

| Variable                | Value                                                                                                                           |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `.chezmoi.arch`         | Architecture, e.g. `amd64`, `arm`, etc. as returned by [runtime.GOARCH](https://pkg.go.dev/runtime?tab=doc#pkg-constants).      |
| `.chezmoi.fqdnHostname` | The fully-qualified domain name hostname of the machine chezmoi is running on.                                                  |
| `.chezmoi.group`        | The group of the user running chezmoi.                                                                                          |
| `.chezmoi.homeDir`      | The home directory of the user running chezmoi.                                                                                 |
| `.chezmoi.hostname`     | The hostname of the machine chezmoi is running on, up to the first `.`.                                                         |
| `.chezmoi.kernel`       | Contains information from `/proc/sys/kernel`. Linux only, useful for detecting specific kernels (i.e. Microsoft's WSL kernel).  |
| `.chezmoi.os`           | Operating system, e.g. `darwin`, `linux`, etc. as returned by [runtime.GOOS](https://pkg.go.dev/runtime?tab=doc#pkg-constants). |
| `.chezmoi.osRelease`    | The information from `/etc/os-release`, Linux only, run `chezmoi data` to see its output.                                       |
| `.chezmoi.sourceDir`    | The source directory.                                                                                                           |
| `.chezmoi.sourceFile`   | The path of the template relative to the source directory.                                                                      |
| `.chezmoi.username`     | The username of the user running chezmoi.                                                                                       |
| `.chezmoi.version`      | The version of chezmoi.                                                                                                         |

Additional variables can be defined in the config file in the `data` section.
Variable names must consist of a letter and be followed by zero or more letters
and/or digits.

---

## Template functions

All standard [`text/template`](https://pkg.go.dev/text/template) and [text
template functions from `sprig`](http://masterminds.github.io/sprig/) are
included. chezmoi provides some additional functions.

---

### `bitwarden` [*arg*...]

`bitwarden` returns structured data retrieved from
[Bitwarden](https://bitwarden.com) using the [Bitwarden
CLI](https://github.com/bitwarden/cli) (`bw`). *arg*s are passed to `bw get`
unchanged and the output from `bw get` is parsed as JSON. The output from `bw
get` is cached so calling `bitwarden` multiple times with the same arguments
will only invoke `bw` once.

#### `bitwarden` examples

```
username = {{ (bitwarden "item" "<itemid>").login.username }}
password = {{ (bitwarden "item" "<itemid>").login.password }}
```

---

### `bitwardenAttachment` *filename* *itemid*

`bitwardenAttachment` returns a document from
[Bitwarden](https://bitwarden.com/) using the [Bitwarden
CLI](https://bitwarden.com/help/article/cli/) (`bw`). *filename* and *itemid* is
passed to `bw get attachment <filename> --itemid <itemid>` and the output from
`bw` is returned. The output from `bw` is cached so calling
`bitwardenAttachment` multiple times with the same *filename* and *itemid* will
only invoke `bw` once.

#### `bitwardenAttachment` examples

```
{{- (bitwardenAttachment "<filename>" "<itemid>") -}}
```

---

### `bitwardenFields` [*arg*...]

`bitwardenFields` returns structured data retrieved from
[Bitwarden](https://bitwarden.com) using the [Bitwarden
CLI](https://github.com/bitwarden/cli) (`bw`). *arg*s are passed to `bw get`
unchanged, the output from `bw get` is parsed as JSON, and elements of `fields`
are returned as a map indexed by each field's `name`. For example, given the
output from `bw get`:

```json
{
    "object": "item",
    "id": "bf22e4b4-ae4a-4d1c-8c98-ac620004b628",
    "organizationId": null,
    "folderId": null,
    "type": 1,
    "name": "example.com",
    "notes": null,
    "favorite": false,
    "fields": [
        {
            "name": "text",
            "value": "text-value",
            "type": 0
        },
        {
            "name": "hidden",
            "value": "hidden-value",
            "type": 1
        }
    ],
    "login": {
        "username": "username-value",
        "password": "password-value",
        "totp": null,
        "passwordRevisionDate": null
    },
    "collectionIds": [],
    "revisionDate": "2020-10-28T00:21:02.690Z"
}
```

the return value will be the map

```json
{
    "hidden": {
        "name": "hidden",
        "type": 1,
        "value": "hidden-value"
    },
    "token": {
        "name": "token",
        "type": 0,
        "value": "token-value"
    }
}
```

The output from `bw get` is cached so calling `bitwarden` multiple times with
the same arguments will only invoke `bw get` once.

#### `bitwardenFields` examples

```
{{ (bitwardenFields "item" "<itemid>").token.value }}
```

---

### `decrypt` *ciphertext*

`decrypt` decrypts *ciphertext* using chezmoi's configured encryption method.

#### `decrypt` examples

```
{{ joinPath .chezmoi.sourceDir ".ignored-encrypted-file.age" | include | decrypt }}
```

---

### `encrypt` *plaintext*

`encrypt` encrypts *plaintext* using chezmoi's configured encryption method.

---

### `gitHubKeys` *user*

`gitHubKeys` returns *user*'s public SSH keys from GitHub using the GitHub API.
The returned value is a slice of structs with `.ID` and `.Key` fields.

**WARNING** if you use this function to populate your `~/.ssh/authorized_keys`
file then you potentially open SSH access to anyone who is able to modify or add
to your GitHub public SSH keys, possibly including certain GitHub employees. You
should not use this function on publicly-accessible machines and should always
verify that no unwanted keys have been added, for example by using the `-v` /
`--verbose` option when running `chezmoi apply` or `chezmoi update`.

By default, an anonymous GitHub API request will be made, which is subject to
[GitHub's rate
limits](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)
(currently 60 requests per hour per source IP address). If any of the
environment variables `$CHEZMOI_GITHUB_ACCESS_TOKEN`, `$GITHUB_ACCESS_TOKEN`, or
`$GITHUB_TOKEN` are found, then the first one found will be used to authenticate
the GitHub API request, with a higher rate limit (currently 5,000 requests per
hour per user).

In practice, GitHub API rate limits are high enough that you should never need
to set a token, unless you are sharing a source IP address with many other
GitHub users. If needed, the GitHub documentation describes how to [create a
personal access
token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).

#### `gitHubKeys` examples

```
{{ range (gitHubKeys "user") }}
{{- .Key }}
{{ end }}
```

---

### `gopass` *gopass-name*

`gopass` returns passwords stored in [gopass](https://www.gopass.pw/) using the
gopass CLI (`gopass`). *gopass-name* is passed to `gopass show --password
<gopass-name>` and the first line of the output of `gopass` is returned with the
trailing newline stripped. The output from `gopass` is cached so calling
`gopass` multiple times with the same *gopass-name* will only invoke `gopass`
once.

#### `gopass` examples

```
{{ gopass "<pass-name>" }}
```

---

### `gopassRaw` *gopass-name*

`gopass` returns passwords stored in [gopass](https://www.gopass.pw/) using the
gopass CLI (`gopass`). *gopass-name* is passed to `gopass show <gopass-name>`
and  output of `gopass` is returned. The output from `gopassRaw` is cached so
calling `gopassRaw` multiple times with the same *gopass-name* will only invoke
`gopass` once.

---

### `include` *filename*

`include` returns the literal contents of the file named `*filename*`. Relative
paths are interpreted relative to the source directory.

---

### `ioreg`

On macOS, `ioreg` returns the structured output of the `ioreg -a -l` command,
which includes detailed information about the I/O Kit registry.

On non-macOS operating systems, `ioreg` returns `nil`.

The output from `ioreg` is cached so multiple calls to the `ioreg` function will
only execute the `ioreg -a -l` command once.

#### `ioreg` examples

```
{{ if (eq .chezmoi.os "darwin") }}
{{ $serialNumber := index ioreg "IORegistryEntryChildren" 0 "IOPlatformSerialNumber" }}
{{ end }}
```

---

### `joinPath` *element*...

`joinPath` joins any number of path elements into a single path, separating them
with the OS-specific path separator. Empty elements are ignored. The result is
cleaned. If the argument list is empty or all its elements are empty, `joinPath`
returns an empty string. On Windows, the result will only be a UNC path if the
first non-empty element is a UNC path.

#### `joinPath` examples

```
{{ joinPath .chezmoi.homeDir ".zshrc" }}
```

---

### `keepassxc` *entry*

`keepassxc` returns structured data retrieved from a
[KeePassXC](https://keepassxc.org/) database using the KeePassXC CLI
(`keepassxc-cli`). The database is configured by setting `keepassxc.database` in
the configuration file. *database* and *entry* are passed to `keepassxc-cli
show`. You will be prompted for the database password the first time
`keepassxc-cli` is run, and the password is cached, in plain text, in memory
until chezmoi terminates. The output from `keepassxc-cli` is parsed into
key-value pairs and cached so calling `keepassxc` multiple times with the same
*entry* will only invoke `keepassxc-cli` once.

#### `keepassxc` examples

```
username = {{ (keepassxc "example.com").UserName }}
password = {{ (keepassxc "example.com").Password }}
```

---

### `keepassxcAttribute` *entry* *attribute*

`keepassxcAttribute` returns the attribute *attribute* of *entry* using
`keepassxc-cli`, with any leading or trailing whitespace removed. It behaves
identically to the `keepassxc` function in terms of configuration, password
prompting, password storage, and result caching.

#### `keepassxcAttribute` examples

```
{{ keepassxcAttribute "SSH Key" "private-key" }}
```

---

### `keyring` *service* *user*

`keyring` retrieves the value associated with *service* and *user* from the
user's keyring.

| OS      | Keyring                     |
| ------- | --------------------------- |
| macOS   | Keychain                    |
| Linux   | GNOME Keyring               |
| Windows | Windows Credentials Manager |

#### `keyring` examples

```
[github]
    user = {{ .github.user | quote }}
    token = {{ keyring "github" .github.user | quote }}
```

---

### `lastpass` *id*

`lastpass` returns structured data from [LastPass](https://lastpass.com) using
the [LastPass CLI](https://lastpass.github.io/lastpass-cli/lpass.1.html)
(`lpass`). *id* is passed to `lpass show --json <id>` and the output from
`lpass` is parsed as JSON. In addition, the `note` field, if present, is further
parsed as colon-separated key-value pairs. The structured data is an array so
typically the `index` function is used to extract the first item. The output
from `lastpass` is cached so calling `lastpass` multiple times with the same
*id* will only invoke `lpass` once.

#### `lastpass` examples

```
githubPassword = {{ (index (lastpass "GitHub") 0).password | quote }}
{{ (index (lastpass "SSH") 0).note.privateKey }}
```

---

### `lastpassRaw` *id*

`lastpassRaw` returns structured data from [LastPass](https://lastpass.com)
using the [LastPass CLI](https://lastpass.github.io/lastpass-cli/lpass.1.html)
(`lpass`). It behaves identically to the `lastpass` function, except that no
further parsing is done on the `note` field.

#### `lastpassRaw` examples

```
{{ (index (lastpassRaw "SSH Private Key") 0).note }}
```

---

### `lookPath` *file*

`lookPath` searches for an executable named *file* in the directories named by
the `PATH` environment variable. If file contains a slash, it is tried directly
and the `PATH` is not consulted. The result may be an absolute path or a path
relative to the current directory. If *file* is not found, `lookPath` returns an
empty string.

`lookPath` is not hermetic: its return value depends on the state of the
environment and the filesystem at the moment the template is executed. Exercise
caution when using it in your templates.

#### `lookPath` examples

```
{{ if lookPath "diff-so-fancy" }}
# diff-so-fancy is in $PATH
{{ end }}
```

---

### `mozillaInstallHash` *path*

`mozillaInstallHash` returns the Mozilla install hash for *path*. This is a
convenience function to assist the management of Firefox profiles.

---

### `onepassword` *uuid* [*vault-uuid* [*account-name*]]

`onepassword` returns structured data from [1Password](https://1password.com/)
using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get item <uuid>` and the output from `op` is parsed as JSON.
The output from `op` is cached so calling `onepassword` multiple times with the
same *uuid* will only invoke `op` once.  If the optional *vault-uuid* is
supplied, it will be passed along to the `op get` call, which can significantly
improve performance. If the optional *account-name* is supplied, it will be
passed along to the `op get` call, which will help it look in the right account,
in case you have multiple accounts (eg. personal and work accounts).

#### `onepassword` examples

```
{{ (onepassword "<uuid>").details.password }}
{{ (onepassword "<uuid>" "<vault-uuid>").details.password }}
{{ (onepassword "<uuid>" "<vault-uuid>" "<account-name>").details.password }}
{{ (onepassword "<uuid>" "" "<account-name>").details.password }}
```

---

### `onepasswordDocument` *uuid* [*vault-uuid* [*account-name*]]

`onepassword` returns a document from [1Password](https://1password.com/) using
the [1Password CLI](https://support.1password.com/command-line-getting-started/)
(`op`). *uuid* is passed to `op get document <uuid>` and the output from `op` is
returned. The output from `op` is cached so calling `onepasswordDocument`
multiple times with the same *uuid* will only invoke `op` once.  If the optional
*vault-uuid* is supplied, it will be passed along to the `op get` call, which
can significantly improve performance. If the optional *account-name* is
supplied, it will be passed along to the `op get` call, which will help it look
in the right account, in case you have multiple accounts (eg. personal and work
accounts).

#### `onepasswordDocument` examples

```
{{- onepasswordDocument "<uuid>" -}}
{{- onepasswordDocument "<uuid>" "<vault-uuid>" -}}
{{- onepasswordDocument "<uuid>" "<vault-uuid>" "<account-name>" -}}
{{- onepasswordDocument "<uuid>" "" "<account-name>" -}}
```

---

### `onepasswordDetailsFields` *uuid* [*vault-uuid* [*account-name*]]

`onepasswordDetailsFields` returns structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get item <uuid>`, the output from `op` is parsed as JSON, and
elements of `details.fields` are returned as a map indexed by each field's
`designation`. For example, give the output from `op`:

```json
{
    "uuid": "<uuid>",
    "details": {
        "fields": [
            {
                "designation": "username",
                "name": "username",
                "type": "T",
                "value": "exampleuser"
            },
            {
                "designation": "password",
                "name": "password",
                "type": "P",
                "value": "examplepassword"
            }
        ]
    }
}
```

the return value will be the map:

```json
{
    "username": {
        "designation": "username",
        "name": "username",
        "type": "T",
        "value": "exampleuser"
    },
    "password": {
        "designation": "password",
        "name": "password",
        "type": "P",
        "value": "examplepassword"
    }
}
```

The output from `op` is cached so calling `onepasswordDetailsFields` multiple
times with the same *uuid* will only invoke `op` once.  If the optional
*vault-uuid* is supplied, it will be passed along to the `op get` call, which
can significantly improve performance. If the optional *account-name* is
supplied, it will be passed along to the `op get` call, which will help it look
in the right account, in case you have multiple accounts (eg. personal and work
accounts).

#### `onepasswordDetailsFields` examples

```
{{ (onepasswordDetailsFields "<uuid>").password.value }}
{{ (onepasswordDetailsFields "<uuid>" "<vault-uuid>").password.value }}
{{ (onepasswordDetailsFields "<uuid>" "<vault-uuid>" "<account-name>").password.value }}
{{ (onepasswordDetailsFields "<uuid>" "" "<account-name>").password.value }}
```

---

### `onepasswordItemFields` *uuid* [*vault-uuid* [*account-name*]]

`onepasswordItemFields` returns structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get item <uuid>`, the output from `op` is parsed as JSON, and
each element of `details.sections` are iterated over and any `fields` are
returned as a map indexed by each field's `n`. For example, give the output from
`op`:

```json
{
  "uuid": "<uuid>",
  "details": {
    "sections": [
      {
        "name": "linked items",
        "title": "Related Items"
      },
      {
        "fields": [
          {
            "k": "string",
            "n": "D4328E0846D2461E8E455D7A07B93397",
            "t": "exampleLabel",
            "v": "exampleValue"
          }
        ],
        "name": "Section_20E0BD380789477D8904F830BFE8A121",
        "title": ""
      }
    ]
  },
}
```

the return value will be the map:

```json
{
    "exampleLabel": {
        "k": "string",
        "n": "D4328E0846D2461E8E455D7A07B93397",
        "t": "exampleLabel",
        "v": "exampleValue"
    }
}
```

The result of

```
{{ (onepasswordItemFields "abcdefghijklmnopqrstuvwxyz").exampleLabel.v }}
```

is equivalent to calling

```console
$ op get item abcdefghijklmnopqrstuvwxyz --fields exampleLabel
```

---

### `output` *name* [*arg*...]

`output` returns the output of executing the command *name* with *arg*s. If
executing the command returns an error then template execution exits with an
error. The execution occurs every time that the template is executed. It is the
user's responsibility to ensure that executing the command is both idempotent
and fast.

#### `output` examples

```
current-context: {{ output "kubectl" "config" "current-context" | trim }}
```

---

### `pass` *pass-name*

`pass` returns passwords stored in [pass](https://www.passwordstore.org/) using
the pass CLI (`pass`). *pass-name* is passed to `pass show <pass-name>` and the
first line of the output of `pass` is returned with the trailing newline
stripped. The output from `pass` is cached so calling `pass` multiple times with
the same *pass-name* will only invoke `pass` once.

#### `pass` examples

```
{{ pass "<pass-name>" }}
```

---

### `passRaw` *pass-name*

`passRaw` returns passwords stored in [pass](https://www.passwordstore.org/)
using the pass CLI (`pass`). *pass-name* is passed to `pass show <pass-name>`
and the output is returned. The output from `pass` is cached so calling
`passRaw` multiple times with the same *pass-name* will only invoke `pass` once.

---

### `promptBool` *prompt* [*default*]

`promptBool` prompts the user with *prompt* and returns the user's response
interpreted as a boolean. If *default* is passed the user's response is empty
then it returns *default*. It is only available when generating the initial
config file. The user's response is interpreted as follows (case insensitive):

| Response                | Result  |
| ----------------------- | ------- |
| 1, on, t, true, y, yes  | `true`  |
| 0, off, f, false, n, no | `false` |

---

### `promptInt` *prompt* [*default*]

`promptInt` prompts the user with *prompt* and returns the user's response
interpreted as an integer. If *default* is passed and the user's response is
empty then it returns *default*. It is only available when generating the
initial config file.

---

### `promptString` *prompt* [*default*]

`promptString` prompts the user with *prompt* and returns the user's response
with all leading and trailing spaces stripped. If *default* is passed and the
user's response is empty then it returns *default*. It is only available when
generating the initial config file.

#### `promptString` examples

```
{{ $email := promptString "email" -}}
[data]
    email = {{ $email | quote }}
```

---

### `secret` [*arg*...]

`secret` returns the output of the generic secret command defined by the
`secret.command` configuration variable with *arg*s with leading and trailing
whitespace removed. The output is cached so multiple calls to `secret` with the
same *arg*s will only invoke the generic secret command once.

---

### `secretJSON` [*arg*...]

`secretJSON` returns structured data from the generic secret command defined by
the `secret.command` configuration variable with *arg*s. The output is parsed as
JSON. The output is cached so multiple calls to `secret` with the same *args*
will only invoke the generic secret command once.

---

### `stat` *name*

`stat` runs `stat(2)` on *name*. If *name* exists it returns structured data. If
*name* does not exist then it returns a false value. If `stat(2)` returns any
other error then it raises an error. The structured value returned if *name*
exists contains the fields `name`, `size`, `mode`, `perm`, `modTime`, and
`isDir`.

`stat` is not hermetic: its return value depends on the state of the filesystem
at the moment the template is executed. Exercise caution when using it in your
templates.

#### `stat` examples

```
{{ if stat (joinPath .chezmoi.homeDir ".pyenv") }}
# ~/.pyenv exists
{{ end }}
```

---

### `stdinIsATTY`

`stdinIsATTY` returns `true` if chezmoi's standard input is a TTY. It is only
available when generating the initial config file. It is primarily useful for
determining whether `prompt*` functions should be called or default values be
used.

#### `stdinIsATTY` examples

```
{{ $email := "" }}
{{ if stdinIsATTY }}
{{   $email = promptString "email" }}
{{ else }}
{{   $email = "user@example.com" }}
{{ end }}
```

---

### `vault` *key*

`vault` returns structured data from [Vault](https://www.vaultproject.io/) using
the [Vault CLI](https://www.vaultproject.io/docs/commands/) (`vault`). *key* is
passed to `vault kv get -format=json <key>` and the output from `vault` is
parsed as JSON. The output from `vault` is cached so calling `vault` multiple
times with the same *key* will only invoke `vault` once.

#### `vault` examples

```
{{ (vault "<key>").data.data.password }}
```

---

### `writeToStdout` *string*...

`writeToStdout` writes each *string* to stdout. It is only available when
generating the initial config file.

#### `writeToStdout` examples

```
{{- writeToStdout "Hello, world\n" -}}
```
