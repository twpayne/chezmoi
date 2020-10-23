# chezmoi Reference Manual

Manage your dotfiles securely across multiple machines.

<!--- toc --->
* [Concepts](#concepts)
* [Global command line flags](#global-command-line-flags)
  * [`--color` *value*](#--color-value)
  * [`-c`, `--config` *filename*](#-c---config-filename)
  * [`--debug`](#--debug)
  * [`-D`, `--destination` *directory*](#-d---destination-directory)
  * [`--follow`](#--follow)
  * [`-n`, `--dry-run`](#-n---dry-run)
  * [`-h`, `--help`](#-h---help)
  * [`-r`. `--remove`](#-r---remove)
  * [`-S`, `--source` *directory*](#-s---source-directory)
  * [`-v`, `--verbose`](#-v---verbose)
  * [`--version`](#--version)
* [Configuration file](#configuration-file)
  * [Variables](#variables)
  * [Examples](#examples)
* [Source state attributes](#source-state-attributes)
* [Special files and directories](#special-files-and-directories)
  * [`.chezmoi.<format>.tmpl`](#chezmoiformattmpl)
  * [`.chezmoiignore`](#chezmoiignore)
  * [`.chezmoiremove`](#chezmoiremove)
  * [`.chezmoitemplates`](#chezmoitemplates)
  * [`.chezmoiversion`](#chezmoiversion)
* [Commands](#commands)
  * [`add` *targets*](#add-targets)
  * [`apply` [*targets*]](#apply-targets)
  * [`archive`](#archive)
  * [`cat` *targets*](#cat-targets)
  * [`cd`](#cd)
  * [`chattr` *attributes* *targets*](#chattr-attributes-targets)
  * [`completion` *shell*](#completion-shell)
  * [`data`](#data)
  * [`diff` [*targets*]](#diff-targets)
  * [`docs` [*regexp*]](#docs-regexp)
  * [`doctor`](#doctor)
  * [`dump` [*targets*]](#dump-targets)
  * [`edit` [*targets*]](#edit-targets)
  * [`edit-config`](#edit-config)
  * [`execute-template` [*templates*]](#execute-template-templates)
  * [`forget` *targets*](#forget-targets)
  * [`git` [*arguments*]](#git-arguments)
  * [`help` *command*](#help-command)
  * [`hg` [*arguments*]](#hg-arguments)
  * [`init` [*repo*]](#init-repo)
  * [`import` *filename*](#import-filename)
  * [`manage` *targets*](#manage-targets)
  * [`managed`](#managed)
  * [`merge` *targets*](#merge-targets)
  * [`purge`](#purge)
  * [`remove` *targets*](#remove-targets)
  * [`rm` *targets*](#rm-targets)
  * [`secret`](#secret)
  * [`source` [*args*]](#source-args)
  * [`source-path` [*targets*]](#source-path-targets)
  * [`unmanage` *targets*](#unmanage-targets)
  * [`unmanaged`](#unmanaged)
  * [`update`](#update)
  * [`upgrade`](#upgrade)
  * [`verify` [*targets*]](#verify-targets)
* [Editor configuration](#editor-configuration)
* [Umask configuration](#umask-configuration)
* [Template execution](#template-execution)
* [Template variables](#template-variables)
* [Template functions](#template-functions)
  * [`bitwarden` [*args*]](#bitwarden-args)
  * [`gopass` *gopass-name*](#gopass-gopass-name)
  * [`include` *filename*](#include-filename)
  * [`ioreg`](#ioreg)
  * [`joinPath` *elements*](#joinpath-elements)
  * [`keepassxc` *entry*](#keepassxc-entry)
  * [`keepassxcAttribute` *entry* *attribute*](#keepassxcattribute-entry-attribute)
  * [`keyring` *service* *user*](#keyring-service-user)
  * [`lastpass` *id*](#lastpass-id)
  * [`lastpassRaw` *id*](#lastpassraw-id)
  * [`lookPath` *file*](#lookpath-file)
  * [`onepassword` *uuid* [*vault-uuid*]](#onepassword-uuid-vault-uuid)
  * [`onepasswordDocument` *uuid* [*vault-uuid*]](#onepassworddocument-uuid-vault-uuid)
  * [`onepasswordDetailsFields` *uuid* [*vault-uuid*]](#onepassworddetailsfields-uuid-vault-uuid)
  * [`pass` *pass-name*](#pass-pass-name)
  * [`promptString` *prompt*](#promptstring-prompt)
  * [`secret` [*args*]](#secret-args)
  * [`secretJSON` [*args*]](#secretjson-args)
  * [`stat` *name*](#stat-name)
  * [`vault` *key*](#vault-key)

## Concepts

chezmoi evaluates the source state for the current machine and then updates the
destination directory, where:

* The *source state* declares the desired state of your home directory,
  including templates and machine-specific configuration.

* The *source directory* is where chezmoi stores the source state, by default
  `~/.local/share/chezmoi`.

* The *target state* is the source state computed for the current machine.

* The *destination directory* is the directory that chezmoi manages, by default
  `~`, your home directory.

* A *target* is a file, directory, or symlink in the destination directory.

* The *destination state* is the current state of all the targets in the
  destination directory.

* The *config file* contains machine-specific configuration, by default it is
  `~/.config/chezmoi/chezmoi.toml`.

## Global command line flags

Command line flags override any values set in the configuration file.

### `--color` *value*

Colorize diffs, *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by
[`strconv.ParseBool`](https://pkg.go.dev/strconv?tab=doc#ParseBool). The default
value is `auto` which will colorize diffs only if the the environment variable
`NO_COLOR` is not set and stdout is a terminal.

### `-c`, `--config` *filename*

Read the configuration from *filename*.

### `--debug`

Log information helpful for debugging.

### `-D`, `--destination` *directory*

Use *directory* as the destination directory.

### `--follow`

If the last part of a target is a symlink, deal with what the symlink
references, rather than the symlink itself.

### `-n`, `--dry-run`

Set dry run mode. In dry run mode, the destination directory is never modified.
This is most useful in combination with the `-v` (verbose) flag to print changes
that would be made without making them.

### `-h`, `--help`

Print help.

### `-r`. `--remove`

Also remove targets according to `.chezmoiremove`.

### `-S`, `--source` *directory*

Use *directory* as the source directory.

### `-v`, `--verbose`

Set verbose mode. In verbose mode, chezmoi prints the changes that it is making
as approximate shell commands, and any differences in files between the target
state and the destination set are printed as unified diffs.

### `--version`

Print the version of chezmoi, the commit at which it was built, and the build
timestamp.

## Configuration file

chezmoi searches for its configuration file according to the [XDG Base Directory
Specification](https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html)
and supports all formats supported by
[`github.com/spf13/viper`](https://github.com/spf13/viper), namely
[JSON](https://www.json.org/json-en.html),
[TOML](https://github.com/toml-lang/toml), [YAML](https://yaml.org/), macOS
property file format, and [HCL](https://github.com/hashicorp/hcl). The basename
of the config file is `chezmoi`, and the first config file found is used.

### Variables

The following configuration variables are available:

| Section         | Variable     | Type     | Default value             | Description                                         |
| --------------- | ------------ | -------- | ------------------------- | --------------------------------------------------- |
| Top level       | `color`      | string   | `auto`                    | Colorize diffs                                      |
|                 | `data`       | any      | *none*                    | Template data                                       |
|                 | `destDir`    | string   | `~`                       | Destination directory                               |
|                 | `dryRun`     | bool     | `false`                   | Dry run mode                                        |
|                 | `follow`     | bool     | `false`                   | Follow symlinks                                     |
|                 | `remove`     | bool     | `false`                   | Remove targets                                      |
|                 | `sourceDir`  | string   | `~/.local/share/chezmoi`  | Source directory                                    |
|                 | `umask`      | int      | *from system*             | Umask                                               |
|                 | `verbose`    | bool     | `false`                   | Verbose mode                                        |
| `bitwarden`     | `command`    | string   | `bw`                      | Bitwarden CLI command                               |
| `cd`            | `args`       | []string | *none*                    | Extra args to shell in `cd` command                 |
|                 | `command`    | string   | *none*                    | Shell to run in `cd` command                        |
| `diff`          | `format`     | string   | `chezmoi`                 | Diff format, either `chezmoi` or `git`              |
|                 | `pager`      | string   | *none*                    | Pager                                               |
| `genericSecret` | `command`    | string   | *none*                    | Generic secret command                              |
| `gopass`        | `command`    | string   | `gopass`                  | gopass CLI command                                  |
| `gpg`           | `command`    | string   | `gpg`                     | GPG CLI command                                     |
|                 | `recipient`  | string   | *none*                    | GPG recipient                                       |
|                 | `symmetric`  | bool     | `false`                   | Use symmetric GPG encryption                        |
| `keepassxc`     | `args`       | []string | *none*                    | Extra args to KeePassXC CLI command                 |
|                 | `command`    | string   | `keepassxc-cli`           | KeePassXC CLI command                               |
|                 | `database`   | string   | *none*                    | KeePassXC database                                  |
| `lastpass`      | `command`    | string   | `lpass`                   | Lastpass CLI command                                |
| `merge`         | `args`       | []string | *none*                    | Extra args to 3-way merge command                   |
|                 | `command`    | string   | `vimdiff`                 | 3-way merge command                                 |
| `onepassword`   | `command`    | string   | `op`                      | 1Password CLI command                               |
| `pass`          | `command`    | string   | `pass`                    | Pass CLI command                                    |
| `sourceVCS`     | `autoCommit` | bool     | `false`                   | Commit changes to the source state after any change |
|                 | `autoPush`   | bool     | `false`                   | Push changes to the source state after any change   |
|                 | `command`    | string   | `git`                     | Source version control system                       |
| `template`      | `options`    | []string | `["missingkey=error"]`    | Template options                                    |
| `vault`         | `command`    | string   | `vault`                   | Vault CLI command                                   |

### Examples

#### JSON

```json
{
    "sourceDir": "/home/user/.dotfiles",
    "diff": {
        "format": "git"
    }
}
```

#### TOML

```toml
sourceDir = "/home/user/.dotfiles"
[diff]
    format = "git"
```

#### YAML

```yaml
sourceDir: /home/user/.dotfiles
diff:
    format: git
```

## Source state attributes

chezmoi stores the source state of files, symbolic links, and directories in
regular files and directories in the source directory (`~/.local/share/chezmoi`
by default). This location can be overridden with the `-S` flag or by giving a
value for `sourceDir` in `~/.config/chezmoi/chezmoi.toml`.  Some state is
encoded in the source names. chezmoi ignores all files and directories in the
source directory that begin with a `.`. The following prefixes and suffixes are
special, and are collectively referred to as "attributes":

| Prefix       | Effect                                                                         |
| ------------ | ------------------------------------------------------------------------------ |
| `encrypted_` | Encrypt the file in the source state.                                          |
| `once_`      | Only run script once.                                                          |
| `private_`   | Remove all group and world permissions from the target file or directory.      |
| `empty_`     | Ensure the file exists, even if is empty. By default, empty files are removed. |
| `exact_`     | Remove anything not managed by chezmoi.                                        |
| `executable_`| Add executable permissions to the target file.                                 |
| `run_`       | Treat the contents as a script to run.                                         |
| `symlink_`   | Create a symlink instead of a regular file.                                    |
| `dot_`       | Rename to use a leading dot, e.g. `dot_foo` becomes `.foo`.                    |

| Suffix  | Effect                                               |
| ------- | ---------------------------------------------------- |
| `.tmpl` | Treat the contents of the source file as a template. |

Order of prefixes is important, the order is `run_`, `exact_`, `private_`,
`empty_`, `executable_`, `symlink_`, `once_`, `dot_`.

Different target types allow different prefixes and suffixes:

| Target type   | Allowed prefixes                                          | Allowed suffixes |
| ------------- | --------------------------------------------------------- | ---------------- |
| Directory     | `exact_`, `private_`, `dot_`                              | *none*           |
| Regular file  | `encrypted_`, `private_`, `empty_`, `executable_`, `dot_` | `.tmpl`          |
| Script        | `run_`, `once_`                                           | `.tmpl`          |
| Symbolic link | `symlink_`, `dot_`,                                       | `.tmpl`          |

## Special files and directories

All files and directories in the source state whose name begins with `.` are
ignored by default, unless they are one of the special files listed here.

### `.chezmoi.<format>.tmpl`

If a file called `.chezmoi.<format>.tmpl` exists then `chezmoi init` will use it
to create an initial config file. *format* must be one of the the supported
config file formats.

#### `.chezmoi.<format>.tmpl` examples

    {{ $email := promptString "email" -}}
    data:
        email: "{{ $email }}"

### `.chezmoiignore`

If a file called `.chezmoiignore` exists in the source state then it is
interpreted as a set of patterns to ignore. Patterns are matched using
[`doublestar.PathMatch`](https://pkg.go.dev/github.com/bmatcuk/doublestar?tab=doc#PathMatch)
and match against the target path, not the source path.

Patterns can be excluded by prefixing them with a `!` character. All excludes
take priority over all includes.

Comments are introduced with the `#` character and run until the end of the
line.

`.chezmoiignore` is interpreted as a template. This allows different files to be
ignored on different machines.

`.chezmoiignore` files in subdirectories apply only to that subdirectory.

#### `.chezmoiignore` examples

    README.md

    *.txt   # ignore *.txt in the target directory
    */*.txt # ignore *.txt in subdirectories of the target directory
    backups/** # ignore backups folder in chezmoi directory and all its contents

    {{- if ne .email "john.smith@company.com" }}
    # Ignore .company-directory unless configured with a company email
    .company-directory # note that the pattern is not dot_company-directory
    {{- end }}

    {{- if ne .email "john@home.org }}
    .personal-file
    {{- end }}

### `.chezmoiremove`

If a file called `.chezmoiremove` exists in the source state then it is
interpreted as a list of targets to remove. `.chezmoiremove` is interpreted as a
template.

### `.chezmoitemplates`

If a directory called `.chezmoitemplates` exists, then all files in this
directory are parsed as templates are available as templates with a name equal
to the relative path of the file.

#### `.chezmoitemplates` examples

Given:

    .chezmoitemplates/foo
    {{ if true }}bar{{ end }}

    dot_config.tmpl
    {{ template "foo" }}

The target state of `.config` will be `bar`.

### `.chezmoiversion`

If a file called `.chezmoiversion` exists, then its contents are interpreted as
a semantic version defining the minimum version of chezmoi required to interpret
the source state correctly. chezmoi will refuse to interpret the source state if
the current version is too old.

**Warning** support for `.chezmoiversion` will be introduced in a future version
(likely 1.5.0). Earlier versions of chezmoi will ignore this file.

#### `.chezmoiversion` examples

    1.5.0

## Commands

### `add` *targets*

Add *targets* to the source state. If any target is already in the source state,
then its source state is replaced with its current state in the destination
directory. The `add` command accepts additional flags:

#### `--autotemplate`

Automatically generate a template by replacing strings with variable names from
the `data` section of the config file. Longer substitutions occur before shorter
ones. This implies the `--template` option.

#### `-e`, `--empty`

Set the `empty` attribute on added files.

#### `-f`, `--force`

Add *targets*, even if doing so would cause a source template to be overwritten.

#### `-x`, `--exact`

Set the `exact` attribute on added directories.

#### `-p`, `--prompt`

Interactively prompt before adding each file.

#### `-r`, `--recursive`

Recursively add all files, directories, and symlinks.

#### `-T`, `--template`

Set the `template` attribute on added files and symlinks.

#### `add` examples

    chezmoi add ~/.bashrc
    chezmoi add ~/.gitconfig --template
    chezmoi add ~/.vim --recursive
    chezmoi add ~/.oh-my-zsh --exact --recursive

### `apply` [*targets*]

Ensure that *targets* are in the target state, updating them if necessary. If no
targets are specified, the state of all targets are ensured.

#### `apply` examples

    chezmoi apply
    chezmoi apply --dry-run --verbose
    chezmoi apply ~/.bashrc

### `archive`

Generate a tar archive of the target state. This can be piped into `tar` to
inspect the target state.

#### `--output`, `-o` *filename*

Write the output to *filename* instead of stdout.

#### `archive` examples

    chezmoi archive | tar tvf -
    chezmoi archive --output=dotfiles.tar

### `cat` *targets*

Write the target state of *targets*  to stdout. *targets* must be files or
symlinks. For files, the target file contents are written. For symlinks, the
target target is written.

#### `cat` examples

    chezmoi cat ~/.bashrc

### `cd`

Launch a shell in the source directory. chezmoi will launch the command set by
the `cd.command` configuration variable with any extra arguments specified by
`cd.args`. If this is not set, chezmoi will attempt to detect your shell and
will finally fall back to an OS-specific default.

#### `cd` examples

    chezmoi cd

### `chattr` *attributes* *targets*

Change the attributes of *targets*. *attributes* specifies which attributes to
modify. Add attributes by specifying them or their abbreviations directly,
optionally prefixed with a plus sign (`+`). Remove attributes by prefixing them
or their attributes with the string `no` or a minus sign (`-`). The available
attributes and their abbreviations are:

| Attribute    | Abbreviation |
| ------------ | ------------ |
| `empty`      | `e`          |
| `encrypted`  | *none*       |
| `exact`      | *none*       |
| `executable` | `x`          |
| `private`    | `p`          |
| `template`   | `t`          |

Multiple attributes modifications may be specified by separating them with a
comma (`,`).

#### `chattr` examples

    chezmoi chattr template ~/.bashrc
    chezmoi chattr noempty ~/.profile
    chezmoi chattr private,template ~/.netrc

### `completion` *shell*

Generate shell completion code for the specified shell (`bash`, `fish`, or
`zsh`).

#### `--output`, `-o` *filename*

Write the shell completion code to *filename* instead of stdout.

#### `completion` examples

    chezmoi completion bash
    chezmoi completion fish --output ~/.config/fish/completions/chezmoi.fish

### `data`

Write the computed template data in JSON format to stdout. The `data` command
accepts additional flags:

#### `-f`, `--format` *format*

Print the computed template data in the given format. The accepted formats are
`json` (JSON), `toml` (TOML), and `yaml` (YAML).

#### `data` examples

    chezmoi data
    chezmoi data --format=yaml

### `diff` [*targets*]

Print the difference between the target state and the destination state for
*targets*. If no targets are specified, print the differences for all targets.

If a `diff.pager` command is set in the configuration file then the output will
be piped into it.

#### `-f`, `--format` *format*

Print the diff in *format*. The format can be set with the `diff.format`
variable in the configuration file. Valid formats are:

##### `chezmoi`

A mix of unified diffs and pseudo shell commands, including scripts, equivalent
to `chezmoi apply --dry-run --verbose`.

##### `git`

A [git format diff](https://git-scm.com/docs/diff-format), excluding scripts. In
version 2.0.0 of chezmoi, `git` format diffs will become the default and include
scripts and the `chezmoi` format will be removed.

#### `--no-pager`

Do not use the pager.

#### `diff` examples

    chezmoi diff
    chezmoi diff ~/.bashrc
    chezmoi diff --format=git

### `docs` [*regexp*]

Print the documentation page matching the regular expression *regexp*. Matching
is case insensitive. If no pattern is given, print `REFERENCE.md`.

#### `docs` examples

    chezmoi docs
    chezmoi docs faq
    chezmoi docs howto

### `doctor`

Check for potential problems.

#### `doctor` examples

    chezmoi doctor

### `dump` [*targets*]

Dump the target state in JSON format. If no targets are specified, then the
entire target state. The `dump` command accepts additional arguments:

#### `-f`, `--format` *format*

Print the target state in the given format. The accepted formats are `json`
(JSON) and `yaml` (YAML).

#### `dump` examples

    chezmoi dump ~/.bashrc
    chezmoi dump --format=yaml

### `edit` [*targets*]

Edit the source state of *targets*, which must be files or symlinks. If no
targets are given the the source directory itself is opened with `$EDITOR`. The
`edit` command accepts additional arguments:

#### `-a`, `--apply`

Apply target immediately after editing. Ignored if there are no targets.

#### `-d`, `--diff`

Print the difference between the target state and the actual state after
editing.. Ignored if there are no targets.

#### `-p`, `--prompt`

Prompt before applying each target.. Ignored if there are no targets.

#### `edit` examples

    chezmoi edit ~/.bashrc
    chezmoi edit ~/.bashrc --apply --prompt
    chezmoi edit

### `edit-config`

Edit the configuration file.

#### `edit-config` examples

    chezmoi edit-config

### `execute-template` [*templates*]

Execute *templates*. This is useful for testing templates or for calling chezmoi
from other scripts. *templates* are interpreted as literal templates, with no
whitespace added to the output between arguments. If no templates are specified,
the template is read from stdin.

#### `--init`, `-i`

Include simulated functions only available during `chezmoi init`.

#### `--output`, `-o` *filename*

Write the output to *filename* instead of stdout.

#### `--promptString`, `-p` *pairs*

Simulate the `promptString` function with a function that returns values from
*pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptString` is called with a *prompt* that does not match any of *pairs*,
then it returns *prompt* unchanged.

#### `execute-template` examples

    chezmoi execute-template '{{ .chezmoi.sourceDir }}'
    chezmoi execute-template '{{ .chezmoi.os }}' / '{{ .chezmoi.arch }}'
    echo '{{ .chezmoi | toJson }}' | chezmoi execute-template
    chezmoi execute-template --init --promptString email=john@home.org < ~/.local/share/chezmoi/.chezmoi.toml.tmpl

### `forget` *targets*

Remove *targets* from the source state, i.e. stop managing them.

#### `forget` examples

    chezmoi forget ~/.bashrc

### `git` [*arguments*]

Run `git` *arguments* in the source directory. Note that flags in *arguments*
must occur after `--` to prevent chezmoi from interpreting them.

#### `git` examples

    chezmoi git add .
    chezmoi git add dot_gitconfig
    chezmoi git -- commit -m "Add .gitconfig"

### `help` *command*

Print the help associated with *command*.

### `hg` [*arguments*]

Run `hg` *arguments* in the source directory. Note that flags in *arguments*
must occur after `--` to prevent chezmoi from interpreting them.

#### `hg` examples

    chezmoi hg -- pull --rebase --update

### `init` [*repo*]

Setup the source directory and update the destination directory to match the
target state.

First, if the source directory is not already contain a repository, then if
*repo* is given it is checked out into the source directory, otherwise a new
repository is initialized in the source directory.

Second, if a file called `.chezmoi.format.tmpl` exists, where `format` is one of
the supported file formats (e.g. `json`, `toml`, or `yaml`) then a new
configuration file is created using that file as a template.

Finally, if the `--apply` flag is passed, `chezmoi apply` is run.

#### `--apply`

Run `chezmoi apply` after checking out the repo and creating the config file.
This is `false` by default.

#### `init` examples

    chezmoi init https://github.com/user/dotfiles.git
    chezmoi init https://github.com/user/dotfiles.git --apply

### `import` *filename*

Import the source state from an archive file in to a directory in the source
state. This is primarily used to make subdirectories of your home directory
exactly match the contents of a downloaded archive. You will generally always
want to set the `--destination`, `--exact`, and `--remove-destination` flags.

The only supported archive format is `.tar.gz`.

#### `--destination` *directory*

Set the destination (in the source state) where the archive will be imported.

#### `-x`, `--exact`

Set the `exact` attribute on all imported directories.

#### `-r`, `--remove-destination`

Remove destination (in the source state) before importing.

#### `--strip-components` *n*

Strip *n* leading components from paths.

#### `import` examples

    curl -s -L -o oh-my-zsh-master.tar.gz https://github.com/robbyrussell/oh-my-zsh/archive/master.tar.gz
    chezmoi import --strip-components 1 --destination ~/.oh-my-zsh oh-my-zsh-master.tar.gz

### `manage` *targets*

`manage` is an alias for `add` for symmetry with `unmanage`.

### `managed`

List all managed entries in the destination directory in alphabetical order.

#### `-i`, `--include` *types*

Only list entries of type *types*. *types* is a comma-separated list of types of
entry to include. Valid types are `dirs`, `files`, and `symlinks` which can be
abbreviated to `d`, `f`, and `s` respectively. By default, `manage` will list
entries of all types.

#### `managed` examples

    chezmoi managed
    chezmoi managed --include=files
    chezmoi managed --include=files,symlinks
    chezmoi managed -i d
    chezmoi managed -i d,f

### `merge` *targets*

Perform a three-way merge between the destination state, the source state, and
the target state. The merge tool is defined by the `merge.command` configuration
variable, and defaults to `vimdiff`. If multiple targets are specified the merge
tool is invoked for each target. If the target state cannot be computed (for
example if source is a template containing errors or an encrypted file that
cannot be decrypted) a two-way merge is performed instead.

#### `merge` examples

    chezmoi merge ~/.bashrc

### `purge`

Remove chezmoi's configuration, state, and source directory, but leave the
target state intact.

#### `-f`, `--force`

Remove without prompting.

#### `purge` examples

    chezmoi purge
    chezmoi purge --force

### `remove` *targets*

Remove *targets* from both the source state and the destination directory.

#### `-f`, `--force`

Remove without prompting.

### `rm` *targets*

`rm` is an alias for `remove`.

### `secret`

Run a secret manager's CLI, passing any extra arguments to the secret manager's
CLI. This is primarily for verifying chezmoi's integration with your secret
manager. Normally you would use template functions to retrieve secrets. Note
that if you want to pass flags to the secret manager's CLI you will need to
separate them with `--` to prevent chezmoi from interpreting them.

To get a full list of available commands run:

    chezmoi secret help

#### `secret` examples

    chezmoi secret bitwarden list items
    chezmoi secret keyring set --service service --user user
    chezmoi secret keyring get --service service --user user
    chezmoi secret lastpass ls
    chezmoi secret lastpass -- show --format=json id
    chezmoi secret onepassword list items
    chezmoi secret onepassword get item id
    chezmoi secret pass show id
    chezmoi secret vault -- kv get -format=json id

### `source` [*args*]

Execute the source version control system in the source directory with *args*.
Note that any flags for the source version control system must be separated with
a `--` to stop chezmoi from reading them.

#### `source` examples

    chezmoi source init
    chezmoi source add .
    chezmoi source commit -- -m "Initial commit"

### `source-path` [*targets*]

Print the path to each target's source state. If no targets are specified then
print the source directory.

#### `source-path` examples

    chezmoi source-path
    chezmoi source-path ~/.bashrc

### `unmanage` *targets*

`unmanage` is an alias for `forget` for symmetry with `manage`.

### `unmanaged`

List all unmanaged files in the destination directory.

#### `unmanaged` examples

    chezmoi unmanaged

### `update`

Pull changes from the source VCS and apply any changes.

#### `update` examples

    chezmoi update

### `upgrade`

Upgrade chezmoi by downloading and installing the latest released version. This
will call the GitHub API to determine if there is a new version of chezmoi
available, and if so, download and attempt to install it in the same way as
chezmoi was previously installed.

If chezmoi was installed with a package manager (`dpkg` or `rpm`) then `upgrade`
will download a new package and install it, using `sudo` if it is installed.
Otherwise, chezmoi will download the latest executable and replace the existing
executable with the new version.

If the `CHEZMOI_GITHUB_API_TOKEN` environment variable is set, then its value
will be used to authenticate requests to the GitHub API, otherwise
unauthenticated requests are used which are subject to stricter [rate
limiting](https://developer.github.com/v3/#rate-limiting). Unauthenticated
requests should be sufficient for most cases.

#### `upgrade` examples

    chezmoi upgrade

### `verify` [*targets*]

Verify that all *targets* match their target state. chezmoi exits with code 0
(success) if all targets match their target state, or 1 (failure) otherwise. If
no targets are specified then all targets are checked.

#### `verify` examples

    chezmoi verify
    chezmoi verify ~/.bashrc

## Editor configuration

The `edit` and `edit-config` commands use the editor specified by the `VISUAL`
environment variable, the `EDITOR` environment variable, or `vi`, whichever is
specified first.

## Umask configuration

By default, chezmoi uses your current umask as set by your operating system and
shell. chezmoi only stores crude permissions in its source state, namely in the
`executable`  and `private` attributes, corresponding to the umasks of `0o111`
and `0o077` respectively.

For machine-specific control of umask, set the `umask` configuration variable in
chezmoi's configuration file, for example:

    umask = 0o22

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

    [template]
      options = ["missingkey=zero"]

For a full list of options, see
[`Template.Option`](https://pkg.go.dev/text/template?tab=doc#Template.Option).

## Template variables

chezmoi provides the following automatically populated variables:

| Variable                | Value                                                                                                                           |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `.chezmoi.arch`         | Architecture, e.g. `amd64`, `arm`, etc. as returned by [runtime.GOARCH](https://pkg.go.dev/runtime?tab=doc#pkg-constants).      |
| `.chezmoi.fullHostname` | The full hostname of the machine chezmoi is running on.                                                                         |
| `.chezmoi.group`        | The group of the user running chezmoi.                                                                                          |
| `.chezmoi.homedir`      | The home directory of the user running chezmoi.                                                                                 |
| `.chezmoi.hostname`     | The hostname of the machine chezmoi is running on, up to the first `.`.                                                         |
| `.chezmoi.kernel`       | Contains information from `/proc/sys/kernel`. Linux only, useful for detecting specific kernels (i.e. Microsoft's WSL kernel).  |
| `.chezmoi.os`           | Operating system, e.g. `darwin`, `linux`, etc. as returned by [runtime.GOOS](https://pkg.go.dev/runtime?tab=doc#pkg-constants). |
| `.chezmoi.osRelease`    | The information from `/etc/os-release`, Linux only, run `chezmoi data` to see its output.                                       |
| `.chezmoi.sourceDir`    | The source directory.                                                                                                           |
| `.chezmoi.username`     | The username of the user running chezmoi.                                                                                       |

Additional variables can be defined in the config file in the `data` section.
Variable names must consist of a letter and be followed by zero or more letters
and/or digits.

## Template functions

All standard [`text/template`](https://pkg.go.dev/text/template) and [text
template functions from `sprig`](http://masterminds.github.io/sprig/) are
included. chezmoi provides some additional functions.

### `bitwarden` [*args*]

`bitwarden` returns structured data retrieved from
[Bitwarden](https://bitwarden.com) using the [Bitwarden
CLI](https://github.com/bitwarden/cli) (`bw`). *args* are passed to `bw`
unchanged and the output from `bw` is parsed as JSON. The output from `bw` is
cached so calling `bitwarden` multiple times with the same arguments will only
invoke `bw` once.

#### `bitwarden` examples

    username = {{ (bitwarden "item" "example.com").login.username }}
    password = {{ (bitwarden "item" "example.com").login.password }}

### `gopass` *gopass-name*

`gopass` returns passwords stored in [gopass](https://www.gopass.pw/) using the
gopass CLI (`gopass`). *gopass-name* is passed to `gopass show <gopass-name>`
and first line of the output of `gopass` is returned with the trailing newline
stripped. The output from `gopass` is cached so calling `gopass` multiple times
with the same *gopass-name* will only invoke `gopass` once.

#### `gopass` examples

    {{ gopass "<pass-name>" }}

### `include` *filename*

`include` returns the literal contents of the file named `*filename*`, relative
to the source directory.

### `ioreg`

On macOS, `ioreg` returns the structured output of the `ioreg -a -l` command,
which includes detailed information about the I/O Kit registry.

On non-macOS operating systems, `ioreg` returns `nil`.

The output from `ioreg` is cached so multiple calls to the `ioreg` function will
only execute the `ioreg -a -l` command once.

#### `ioreg` examples

    {{ if (eq .chezmoi.os "darwin") }}
    {{ $serialNumber := index ioreg "IORegistryEntryChildren" 0 "IOPlatformSerialNumber" }}
    {{ end }}

### `joinPath` *elements*

`joinPath` joins any number of path elements into a single path, separating them
with the OS-specific path separator. Empty elements are ignored. The result is
cleaned. If the argument list is empty or all its elements are empty, `joinPath`
returns an empty string. On Windows, the result will only be a UNC path if the
first non-empty element is a UNC path.

#### `joinPath` examples

    {{ joinPath .chezmoi.homedir ".zshrc" }}

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

    username = {{ (keepassxc "example.com").UserName }}
    password = {{ (keepassxc "example.com").Password }}

### `keepassxcAttribute` *entry* *attribute*

`keepassxcAttribute` returns the attribute *attribute* of *entry* using
`keepassxc-cli`, with any leading or trailing whitespace removed. It behaves
identically to the `keepassxc` function in terms of configuration, password
prompting, password storage, and result caching.

#### `keepassxcAttribute` examples

    {{ keepassxcAttribute "SSH Key" "private-key" }}

### `keyring` *service* *user*

`keyring` retrieves the password associated with *service* and *user* from the
user's keyring.

| OS    | Keyring       |
| ----- | ------------- |
| macOS | Keychain      |
| Linux | GNOME Keyring |

#### `keyring` examples

    [github]
      user = "{{ .github.user }}"
      token = "{{ keyring "github" .github.user }}"

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

    githubPassword = "{{ (index (lastpass "GitHub") 0).password }}"
    {{ (index (lastpass "SSH") 0).note.privateKey }}

### `lastpassRaw` *id*

`lastpassRaw` returns structured data from [LastPass](https://lastpass.com)
using the [LastPass CLI](https://lastpass.github.io/lastpass-cli/lpass.1.html)
(`lpass`). It behaves identically to the `lastpass` function, except that no
further parsing is done on the `note` field.

#### `lastpassRaw` examples

    {{ (index (lastpassRaw "SSH Private Key") 0).note }}

### `lookPath` *file*

`lookPath` searches for an executable named *file* in the directories named by
the `PATH` environment variable. If file contains a slash, it is tried directly
and the `PATH `is not consulted. The result may be an absolute path or a path
relative to the current directory. If *file* is not found, `lookPath` returns an
empty string.

`lookPath` is not hermetic: its return value depends on the state of the
environment and the filesystem at the moment the template is executed. Exercise
caution when using it in your templates.

#### `lookPath` examples

    {{ if lookPath "diff-so-fancy" }}
    # diff-so-fancy is in $PATH
    {{ end }}

### `onepassword` *uuid* [*vault-uuid*]

`onepassword` returns structured data from [1Password](https://1password.com/)
using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get item <uuid>` and the output from `op` is parsed as JSON.
The output from `op` is cached so calling `onepassword` multiple times with the
same *uuid* will only invoke `op` once.  If the optional *vault-uuid* is supplied,
it will be passed along to the `op get` call, which can significantly improve
performance.

#### `onepassword` examples

    {{ (onepassword "<uuid>").details.password }}
    {{ (onepassword "<uuid>" "<vault-uuid>").details.password }}

### `onepasswordDocument` *uuid* [*vault-uuid*]

`onepassword` returns a document from [1Password](https://1password.com/)
using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get document <uuid>` and the output from `op` is returned.
The output from `op` is cached so calling `onepasswordDocument` multiple times with the
same *uuid* will only invoke `op` once.  If the optional *vault-uuid* is supplied,
it will be passed along to the `op get` call, which can significantly improve
performance.

#### `onepasswordDocument` examples

    {{- onepasswordDocument "<uuid>" -}}
    {{- onepasswordDocument "<uuid>" "<vault-uuid>" -}}

### `onepasswordDetailsFields` *uuid* [*vault-uuid*]

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
    ],
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

The output from `op` is cached so calling `onepassword` multiple times with the
same *uuid* will only invoke `op` once.  If the optional *vault-uuid* is supplied,
it will be passed along to the `op get` call, which can significantly improve
performance.

#### `onepasswordDetailsFields` examples

    {{ (onepasswordDetailsFields "<uuid>").password.value }}

### `pass` *pass-name*

`pass` returns passwords stored in [pass](https://www.passwordstore.org/) using
the pass CLI (`pass`). *pass-name* is passed to `pass show <pass-name>` and
first line of the output of `pass` is returned with the trailing newline
stripped. The output from `pass` is cached so calling `pass` multiple times with
the same *pass-name* will only invoke `pass` once.

#### `pass` examples

    {{ pass "<pass-name>" }}

### `promptString` *prompt*

`promptString` takes a single argument is a string prompted to the user, and the
return value is the user's response to that prompt with all leading and trailing
space stripped. It is only available when generating the initial config file.

#### `promptString` examples

    {{ $email := promptString "email" -}}
    [data]
        email = "{{ $email }}"

### `secret` [*args*]

`secret` returns the output of the generic secret command defined by the
`genericSecret.command` configuration variable with *args* with leading and
trailing whitespace removed. The output is cached so multiple calls to `secret`
with the same *args* will only invoke the generic secret command once.

### `secretJSON` [*args*]

`secretJSON` returns structured data from the generic secret command defined by
the `genericSecret.command` configuration variable with *args*. The output is
parsed as JSON. The output is cached so multiple calls to `secret` with the same
*args* will only invoke the generic secret command once.

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

    {{ if stat (joinPath .chezmoi.homedir ".pyenv") }}
    # ~/.pyenv exists
    {{ end }}

### `vault` *key*

`vault` returns structured data from [Vault](https://www.vaultproject.io/) using
the [Vault CLI](https://www.vaultproject.io/docs/commands/) (`vault`). *key* is
passed to `vault kv get -format=json <key>` and the output from `vault` is
parsed as JSON. The output from `vault` is cached so calling `vault` multiple
times with the same *key* will only invoke `vault` once.

#### `vault` examples

    {{ (vault "<key>").data.data.password }}
