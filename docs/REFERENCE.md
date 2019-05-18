# chezmoi Reference Manual

Manage your dotfiles securely across multiple machines.

* [Concepts](#concepts)
* [Global command line flags](#global-command-line-flags)
  * [`--color` *value*](#--color-value)
  * [`-c`, `--config` *filename*](#-c---config-filename)
  * [`-D`, `--destination` *directory*](#-d---destination-directory)
  * [`-n`, `--dry-run`](#-n---dry-run)
  * [`-h`, `--help`](#-h---help)
  * [`-S`, `--source` *directory*](#-s---source-directory)
  * [`-v`, `--verbose`](#-v---verbose)
  * [`--version`](#--version)
* [Configuration file](#configuration-file)
  * [Configuration variables](#configuration-variables)
* [Attributes](#attributes)
* [Special files](#special-files)
  * [`.chezmoi.<format>.tmpl`](#chezmoiformattmpl)
  * [`.chezmoiignore`](#chezmoiignore)
  * [`.chezmoiversion`](#chezmoiversion)
* [Commands](#commands)
  * [`add` *targets*](#add-targets)
  * [`apply` [*targets*]](#apply-targets)
  * [`archive`](#archive)
  * [`cat` targets](#cat-targets)
  * [`cd`](#cd)
  * [`chattr` *attributes* *targets*](#chattr-attributes-targets)
  * [`completion` *shell*](#completion-shell)
  * [`data`](#data)
  * [`diff` [*targets*]](#diff-targets)
  * [`doctor`](#doctor)
  * [`dump` [*targets*]](#dump-targets)
  * [`edit` *targets*](#edit-targets)
  * [`edit-config`](#edit-config)
  * [`forget` *targets*](#forget-targets)
  * [`help` *command*](#help-command)
  * [`init` [*repo*]](#init-repo)
  * [`import` *filename*](#import-filename)
  * [`merge` *targets*](#merge-targets)
  * [`remove` *targets*](#remove-targets)
  * [`secret`](#secret)
  * [`source` [*args*]](#source-args)
  * [`source-path` [*targets*]](#source-path-targets)
  * [`unmanaged`](#unmanaged)
  * [`update`](#update)
  * [`upgrade`](#upgrade)
  * [`verify` [*targets*]](#verify-targets)
* [Editor configuration](#editor-configuration)
* [Umask configuration](#umask-configuration)
* [Template variables](#template-variables)
* [Template functions](#template-functions)
  * [`bitwarden` [*args*]](#bitwarden-args)
  * [`keepassxc` *entry*](#keepassxc-entry)
  * [`keyring` *service* *user*](#keyring-service-user)
  * [`lastpass` *id*](#lastpass-id)
  * [`onepassword` *uuid*](#onepassword-uuid)
  * [`pass` *pass-name*](#pass-pass-name)
  * [`promptString` *prompt*](#promptstring-prompt)
  * [`secret` [*args*]](#secret-args)
  * [`secretJSON` [*args*]](#secretjson-args)
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

* The *destination state* is the state of all the targets in the destination
  directory.

* The *config file* contains machine-specific configuration, by default it is
  `~/.config/chezmoi/chezmoi.toml`.

## Global command line flags

Command line flags override any values set in the configuration file.

### `--color` *value*

Colorize diffs, *value* can be `on`, `off`, or `auto`. The default value is
`auto` which will colorize diffs only if the output is a terminal.

### `-c`, `--config` *filename*

Read the configuration from *filename*.

### `-D`, `--destination` *directory*

Use *directory* as the destination directory.

### `-n`, `--dry-run`

Set dry run mode. In dry run mode, the destination directory is never modified.
This is most useful in combination with the `-v` (verbose) flag to print changes
that would be made without making them.

### `-h`, `--help`

Print help.

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
[`github.com/spf13/viper`](https://github.com/spf13/viper), namely JSON, TOML,
YAML, macOS property file format, and HCL. The basename of the config file is
chezmoi, and the first config file found is used.

### Configuration variables

The following configuration variables are available:

| Variable                | Type     | Default value             | Description                         |
| ----------------------- | -------- | ------------------------- | ----------------------------------- |
| `bitwarden.command`     | string   | `bw`                      | Bitwarden CLI command               |
| `color`                 | string   | `auto`                    | Colorize diffs                      |
| `data`                  | any      | none                      | Template data                       |
| `destDir`               | string   | `~`                       | Destination directory               |
| `dryRun`                | boolean  | `false`                   | Dry run mode                        |
| `genericSecret.command` | string   | none                      | Generic secret command              |
| `gpgRecipient`          | string   | none                      | GPG recipient                       |
| `keepassxc.args`        | []string | none                      | Extra args to KeePassXC CLI command |
| `keepassxc.command`     | string   | `keepassxc-cli`           | KeePassXC CLI command               |
| `keepassxc.database`    | string   | none                      | KeePassXC database                  |
| `lastpass.command`      | string   | `lpass`                   | Lastpass CLI command                |
| `merge.args`            | []string | none                      | Extra args to 3-way merge command   |
| `merge.command`         | string   | `vimdiff`                 | 3-way merge command                 |
| `onepassword.command`   | string   | `op`                      | 1Password CLI command               |
| `pass.command`          | string   | `pass`                    | Pass CLI command                    |
| `sourceDir`             | string   | `~/.config/share/chezmoi` | Source directory                    |
| `sourceVCS.command`     | string   | `git`                     | Source version control system       |
| `umask`                 | integer  | from system               | Umask                               |
| `vault.command`         | string   | `vault`                   | Vault CLI command                   |
| `verbose`               | boolean  | `false`                   | Verbose mode                        |

In addition, a number of secret manager integrations add configuration
variables. These are documented in the secret manager section.

## Attributes

chezmoi stores the source state of files, symbolic links, and directories in
regular files and directories in the source directory (`~/.local/share/chezmoi`
by default). This location can be overridden with the `-S` flag or by giving a
value for `sourceDir` in `~/.config/chezmoi/chezmoi.toml`.  Some state is
encoded in the source names. chezmoi ignores all files and directories in the
source directory that begin with a `.`. The following prefixes and suffixes are
special, and are collectively referred to as "attributes":

| Prefix/suffix        | Effect                                                                            |
| -------------------- | ----------------------------------------------------------------------------------|
| `encrypted_` prefix  | Encrypt the file in the source state.                                             |
| `once_` prefix       | Only run script once.                                                             |
| `private_` prefix    | Remove all group and world permissions from the target file or directory.         |
| `empty_` prefix      | Ensure the file exists, even if is empty. By default, empty files are removed.    |
| `exact_` prefix      | Remove anything not managed by chezmoi.                                           |
| `executable_` prefix | Add executable permissions to the target file.                                    |
| `run_` prefix        | Treat the contents as a script to run.                                            |
| `symlink_` prefix    | Create a symlink instead of a regular file.                                       |
| `dot_` prefix        | Rename to use a leading dot, e.g. `dot_foo` becomes `.foo`.                       |
| `.tmpl` suffix       | Treat the contents of the source file as a template.                              |

Order is important, the order is `run_`, `exact_`, `private_`, `empty_`,
`executable_`, `symlink_`, `once_`, `dot_`, `.tmpl`.

Different target types allow different prefixes and suffixes:

| Target type   | Allowed prefixes and suffixes                                      |
| ------------- | ------------------------------------------------------------------ |
| Directory     | `exact_`, `private_`, `dot_`                                       |
| Regular file  | `encrypted_`, `private_`, `empty_`, `executable_`, `dot_`, `.tmpl` |
| Script        | `run_`, `once_`, `.tmpl`                                           |
| Symbolic link | `symlink_`, `dot_`, `.tmpl`                                        |

## Special files

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
interpreted as a list of globs to ignore. Notably, `.chezmoiignore` is
interpreted as a template. Patterns can be excluded by prefixing them with a
`!`. All excludes (don't ignore) take priority over all includes (ignore).
`.chezmoiignore` files in subdirectories apply only to that subdirectory.

`.chezmoiignore` is inspired by git's `.gitignore` files, but is a separate
implementation and corner case behaviour may differ.

#### `.chezmoiignore` examples

    README.md
    {{- if ne .email "john.smith@company.com" }}
    .company-directory
    {{- end }}
    {{- if ne .email "john@home.org }}
    .personal-file
    {{- end }}

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

#### `-e`, `--empty`

Set the `empty` attribute on added files.

#### `-x`, `--exact`

Set the `exact` attribute on added directories.

#### `-p`, `--prompt`

Interactively prompt before adding each file.

#### `-r`, `--recursive`

Recursively add all files, directories, and symlinks.

#### `-T`, `--template`

Set the `template` attribute on added files and symlinks. In addition, chezmoi
attempts to automatically generate the template by replacing any template data
values with the equivalent template data keys. Longer subsitutions occur before
shorter ones.

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

Write a tar archive of the target state to stdout. This can be piped into `tar`
to inspect the target state.

#### `archive` examples

    chezmoi archive | tar tvf -

### `cat` targets

Write the target state of *targets*  to stdout. *targets* must be files or
symlinks. For files, the target file contents are written. For symlinks, the
target target is written.

#### `cat` examples

    chezmoi cat ~/.bashrc

### `cd`

Launch a shell in the source directory.

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
| `encrypted`  | none         |
| `exact`      | none         |
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

Output shell completion code for the specified shell (`bash` or `zsh`).

#### `completion` examples

    chezmoi completion bash
    chezmoi completion zsh

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

Print the approximate shell commands required to ensure that *targets* in the
destination directory match the target state. If no targets are specifed, print
the commands required for all targets. It is equivalent to `chezmoi apply
--dry-run --verbose`.

#### `diff` examples

    chezmoi diff
    chezmoi diff ~/.bashrc

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

#### `edit` examples

    chezmoi edit ~/.bashrc
    chezmoi edit ~/.bashrc --apply --prompt

### `edit-config`

Edit the configuration file.

#### `edit-config` examples

    chezmoi edit-config

### `forget` *targets*

Remove *targets* from the source state, i.e. stop managing them.

#### `forget` examples

    chezmoi forget ~/.bashrc

### `help` *command*

Print the help associated with *command*.

### `init` [*repo*]

Setup the source directory and update the destination directory to match the
target state. If *repo* is given then it is checked out into the source
directory, otherwise a new repository is initialized in the source directory. If
a file called `.chezmoi.format.tmpl` exists, where `format` is one of the
supported file formats (e.g. `json`, `toml`, or `yaml`) then a new configuration
file is created using that file as a template. Finally, if the `--apply` flag is
passed, `chezmoi apply` is run.

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

### `merge` *targets*

Perform a three-way merge between the destination state, the source state, and
the target state. The merge tool is defined by the `merge.command` configuration
variable, and defaults to `vimdiff`. If multiple targets are specified the merge
tool is invoked for each target. If the target state cannot be computed (for
example if source is a template containing errors or an encrypted file that
cannot be decrypted) a two-way merge is performed instead.

#### `merge` examples

    chezmoi merge ~/.bashrc

### `remove` *targets*

Remove *targets* from both the source state and the destination directory.

### `secret`

Run a secret manager's CLI, passing any extra arguments to the secret manager's
CLI. This is primarily for verifying chezmoi's integration with your secret
manager. Normally you would use template functions to retrieve secrets. Note
that if you want to pass flags to the secret manager's CLU you will need to
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
Note that any flags for the source version control system must be sepeated with
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

### `unmanaged`

List all unmanaged files in the destination directory.

#### `unmanaged` examples

    chezmoi unmanaged

### `update`

Pull changes from the source VCS and apply any changes.

#### `update` examples

    chezmoi update

### `upgrade`

Upgrade chezmoi by downloading and installing a new version. This will call the
Github API to determine if there is a new version of chezmoi available, and if
so, download and attempt to install it in the same way as chezmoi was previously
installed.

If chezmoi was installed with a package manager (`dpkg` or `rpm`) then `upgrade`
will download a new package and install it, using `sudo` if it is installed.
Otherwise, chezmoi will download the latest executable and replace the existing
exectuable with the new version.

If the `CHEZMOI_GITHUB_API_TOKEN` environment variable is set, then its value
will be used to authenticate requests to the Github API, otherwise
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
`executable`  and `private` attributes, corresponding to the umasks of `0111`
and `0077` respectively.

For machine-specifc control of umask, set the `umask` configuration variable in
chezmoi's configuration file.

## Template variables

chezmoi provides the following automatically populated variables:

| Variable                | Value                                                                                                                  |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `.chezmoi.arch`         | Architecture, e.g. `amd64`, `arm`, etc. as returned by [runtime.GOARCH](https://godoc.org/runtime#pkg-constants).      |
| `.chezmoi.fullHostname` | The full hostname of the machine chezmoi is running on.                                                                |
| `.chezmoi.group`        | The group of the user running chezmoi.                                                                                 |
| `.chezmoi.homedir`      | The home directory of the user running chezmoi.                                                                        |
| `.chezmoi.hostname`     | The hostname of the machine chezmoi is running on, up to the first `.`.                                                |
| `.chezmoi.os`           | Operating system, e.g. `darwin`, `linux`, etc. as returned by [runtime.GOOS](https://godoc.org/runtime#pkg-constants). |
| `.chezmoi.osRelease`    | The information from `/etc/os-release`, Linux only, run `chezmoi data` to see its output.                              |
| `.chezmoi.username`     | The username of the user running chezmoi.                                                                              |

Additional variables can be defined in the config file in the `data` section.

## Template functions

All standard [`text/template`](https://godoc.org/text/template) and [hermetic
text template functions from `sprig`](http://masterminds.github.io/sprig/) are
included. chezmoi provides some additional functions.

### `bitwarden` [*args*]

`bitwarden` returns structured data retrieved from
[Bitwarden](https://bitwarden.com) using the [Bitwarden
CLI](https://github.com/bitwarden/cli) (`bw`). *args* are passed to `bw`
unchanged and the output from `bw` is parsed as JSON. The output from `bw` is
cached so calling `bitwarden` multitple times with the same arguments will only
invoke `bw` once.

#### `bitwarden` examples

    username = {{ (bitwarden "item" "example.com").login.username }}
    password = {{ (bitwarden "item" "example.com").login.password }}

### `keepassxc` *entry*

`keepassxc` returns structured data retrieved from a
[KeePassXC](https://keepassxc.org/) database using the KeePassXC CLI
(`keepassxc-cli`). The database is configured by setting `keepassxc.database` in
the configuration file. *database* and *entry* are passed to `keepassxc-cli
show`. You will be prompted for the database password the first time
`keepassxc-cli` is run, and the password is cached, in plain text, in memory
until chezmoi terminates. The output from `keepassxc-cli` is parsed into
key-value pairs. The output from `keepassxc-cli` is cached so calling
`keepassxc` multiple times with the same *entry* will only invoke
`keepassxc-cli` once.

#### `keepassxc` examples

    username = {{ (keepassxc "example.com").UserName }}
    password = {{ (keepassxc "example.com").Password }}

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
parsed as JSON. The structured data is an array so typically the `index`
function is used to extract the first item. The output from `lpass` is cached so
calling `lastpass` multiple times with the same *id* will only invoke `lpass`
once.

#### `lastpass` examples

    githubPassword = "{{ (index (lastpass "Github") 0).password }}"
    {{ (index (lastpass "SSH") 0).note.privateKey }}

### `onepassword` *uuid*

`onepassword` returns structured data from [1Password](https://1password.com/)
using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get item <uuid>` and the output from `op` is parsed as JSON.
The output from `op` is cached so calling `onepassword` multiple times with the
same *uuid* will only invoke `op` once.

#### `onepassword` examples

    {{ (onepassword "<uuid>").details.password }}

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
return value is the user's response to that prompt. It is only available when
generating the initial conifig file.

#### `promptString` examples

    {{ $email := promptString "email" -}}
    data:
        email: "{{ $email }}"

### `secret` [*args*]

`secret` returns the output of the generic secret command defined by the
`genericSecret.command` configuration variable with *args* with leading and
trailing whitespace removed. The output is cached so multitple calls to `secret`
with the same *args* will only invoke the generic secret command once.

### `secretJSON` [*args*]

`secretJSON` returns structured data from the generic secret command defined by
the `genericSecret.command` configuration variable with *args*. The output is
parsed as JSON. The output is cached so multitple calls to `secret` with the
same *args* will only invoke the generic secret command once.

### `vault` *key*

`vault` returns structured data from [Vault](https://www.vaultproject.io/) using
the [Vault CLI](https://www.vaultproject.io/docs/commands/) (`vault`). *key* is
passed to `vault kv get -format=json <key>` and the output from `vault` is
parsed as JSON. The output from `vault` is cached so calling `vault` multiple
times with the same *key* will only invoke `vault` once.

#### `vault` examples

    {{ (vault "<key>").data.data.password }}
