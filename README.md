# `chezmoi`

[![Build
Status](https://travis-ci.org/twpayne/chezmoi.svg?branch=master)](https://travis-ci.org/twpayne/chezmoi)
[![GoDoc](https://godoc.org/github.com/twpayne/chezmoi?status.svg)](https://godoc.org/github.com/twpayne/chezmoi)
[![Report
Card](https://goreportcard.com/badge/github.com/twpayne/chezmoi)](https://goreportcard.com/report/github.com/twpayne/chezmoi)
[![Coverage Status](https://coveralls.io/repos/github/twpayne/chezmoi/badge.svg)](https://coveralls.io/github/twpayne/chezmoi)

Manage your dotfiles across multiple machines, securely.

## Features

* Flexible: your dotfiles can be templates (using
  [`text/template`](https://godoc.org/text/template) syntax). Predefined
  variables allow you to change behaviour depending on operating system,
  architecture, and hostname. You can share as much configuration across
  machines as you want, while still being able to control machine-specific
  details.

* Secure: `chezmoi` can retrieve secrets from
  [1Password](https://1password.com/), [Bitwarden](https://bitwarden.com/),
  [LastPass](https://lastpass.com/), [pass](https://www.passwordstore.org/),
  [Vault](https://www.vaultproject.io/), your Keychain (on macOS), [GNOME
  Keyring](https://wiki.gnome.org/Projects/GnomeKeyring) (on Linux), or any
  command-line utility of your choice. You can encrypt individual files with
  [`gpg`](https://www.gnupg.org). You can checkout your dotfiles repo on as many
  machines as you want without revealing any secrets to anyone.

* Personal: Nothing leaves your machine, unless you want it to. You can use the
  version control system of your choice to manage your configuration, and you
  can write the configuration file in the format of your choice.

* Transparent: `chezmoi` includes verbose and dry run modes so you can review
  exactly what changes it will make to your home directory before making them.
  `chezmoi`'s source format uses only regular files and directories that map
  one-to-one with the files, directories, and symlinks in your home directory
  that you choose to manage. If you decide not to use `chezmoi` in the future,
  it is easy to move your data elsewhere.

* Robust: `chezmoi` updates all files and symbolic links atomically (using
  [`google/renameio`](https://github.com/google/renameio)). You will never be
  left with incomplete files that could lock you out, even if the update process
  is interrupted.

* Declarative: you declare the desired state of files, directories, and symbolic
  links in your home directory and `chezmoi` updates your home directory to
  match that state. What you want is what you get.

* Fast and easy to use: `chezmoi` runs in fractions of a second and makes most
  day-to-day operations one line commands.

## I already have a system to manage my dotfiles, why should I use `chezmoi`?

* If your system is based on copying files with a shell script or creating
  symlinks (e.g. using [GNU
  Stow](http://brandon.invergo.net/news/2012-05-26-using-gnu-stow-to-manage-your-dotfiles.html))
  then handling files that vary from machine to machine requires manual work.
  You might need to maintain separate config files for separate machines, or run
  different commands on different machines. `chezmoi` gives you a single command
  that works on every machine.

* If your system is based on using `git` with a different branches for different
  machines, then you need manually merge or rebase to ensure that changes you
  make are applied to each machine. `chezmoi` makes it trivial to share common
  parts while allowing specific per-machine configuration.

* If your system stores secrets in plain text, then your dotfiles repository
  must be private. With `chezmoi` you never need to store secrets in your
  repository, so you can make it public. You can check out your repository on
  your work machine and not fear that this will give your work IT department
  access to your personal data.

* If your system was written by you for your personal use, then it probably has
  the minimum functionality that you need. `chezmoi` includes a wide range of
  functionality out-of-the-box, including dry run and diff modes.

* All systems suffer from the "bootstrap" problem: you need to install your
  system before you can install your dotfiles. `chezmoi` provides
  statically-linked binaries, packages for many Linux and BSD distributions, and
  Homebrew formulae to make overcoming the bootstrap problem as simple as
  possible.

## Installation

Pre-built packages and binaries:

| OS         | Architectures                                 | Package location                                                     |
| ---------- | --------------------------------------------- | -------------------------------------------------------------------- |
| Arch Linux | i686, x86_64                                  | [aur package](https://aur.archlinux.org/packages/chezmoi/)           |
| Debian     | amd64, arm64, armel, i386, ppc64, ppc64le     | [deb package](https://github.com/twpayne/chezmoi/releases/latest)    |
| RedHat     | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [rpm package](https://github.com/twpayne/chezmoi/releases/latest)    |
| OpenSUSE   | aarch64, armhfp, i686, ppc64, ppc64le, x86_64 | [rpm package](https://github.com/twpayne/chezmoi/releases/latest)    |
| FreeBSD    | amd64, arm, i386                              | [tar.gz package](https://github.com/twpayne/chezmoi/releases/latest) |
| OpenBSD    | amd64, arm, i386                              | [tar.gz package](https://github.com/twpayne/chezmoi/releases/latest) |
| Linux      | amd64, arm, arm64, i386, ppc64, ppc64le       | [tar.gz package](https://github.com/twpayne/chezmoi/releases/latest) |

On macOS you can install `chezmoi` with Homebrew:

    brew install twpayne/taps/chezmoi

If you have Go installed you can install the latest version from `HEAD`:

    go get -u github.com/twpayne/chezmoi

## Quick start

`chezmoi` evaluates the source state for the current machine and then updates
the destination directory, where:

* The *source state* declares the desired state of your home directory,
  including templates and machine-specific configuration.

* The *source directory* is where `chezmoi` stores the source state, by default
  `~/.local/share/chezmoi`.

* The *target state* is the source state computed for the current machine.

* The *destination directory* is the directory that `chezmoi` manages, by
  default `~`, your home directory.

* A *target* is a file, directory, or symlink in the destination directory.

* The *destination state* is the state of all the targets in the destination
  directory.

* The *config file* contains machine-specific configuration, by default it is
  `~/.config/chezmoi/chezmoi.toml`.

Initialize `chezmoi`:

    chezmoi init

This will create a new `git` repository in `~/.local/share/chezmoi` with
permissions `0700` where `chezmoi` will store the source state.  `chezmoi`
generally only modifies files in the working copy. It is your responsibility to
commit changes.

`chezmoi` works with many other version control systems too. See "Using non-git
version control systems" below if you don't want to use `git`.

Manage an existing file with `chezmoi`:

    chezmoi add ~/.bashrc

This will copy `~/.bashrc` to `~/.local/share/chezmoi/dot_bashrc`.

Edit the source state:

    chezmoi edit ~/.bashrc

This will open `~/.local/share/chezmoi/dot_bashrc` in your `$EDITOR`. Make some
changes and save them.

See what changes `chezmoi` would make:

    chezmoi diff

Apply the changes:

    chezmoi -v apply

All `chezmoi` commands accept the `-v` (verbose) flag to print out exactly what
changes they will make to the file system, and the `-n` (dry run) flag to not
make any actual changes. The combination `-n` `-v` is very useful if you want to
see exactly what changes would be made.

Finally, change to the source directory, commit your changes, and return to
where you were:

    chezmoi cd
    git add dot_bashrc
    git commit -m "Updated .bashrc"
    exit

For a full list of commands run:

    chezmoi help

## Using a hosted repo to manage your dotfiles across multiple machines

`chezmoi` relies on your version control system and hosted repo to share changes
across multiple machines. You should create a repo on the source code repository
of your choice (e.g. [Bitbucket](https://bitbucket.org),
[Github](https://github.com/), or [GitLab](https://gitlab.com), many people call
their repo `dotfiles`) and push the repo in the source directory here. For
example:

    chezmoi cd
    git remote add origin https://github.com/username/dotfiles.git
    git push -u origin master
    exit

On another machine you can checkout this repo:

    chezmoi init https://github.com/username/dotfiles.git

You can then see what would be changed:

    chezmoi diff

If you're happy with the changes then apply them:

    chezmoi apply

The above commands can be combined into a single init, checkout, and apply:

    chezmoi init --apply --verbose https://github.com/username/dotfiles.git

You can pull the changes from your repo and apply them in a single command:

    chezmoi update

This runs `git pull --rebase` in your source directory and then `chezmoi apply`.

## Using templates to manage files that vary from machine to machine

The primary goal of `chezmoi` is to manage configuration files across multiple
machines, for example your personal macOS laptop, your work Ubuntu desktop, and
your work Linux laptop. You will want to keep much configuration the same across
these, but also need machine-specific configurations for email addresses,
credentials, etc. `chezmoi` achieves this functionality by using
[`text/template`](https://godoc.org/text/template) for the source state where
needed.

For example, your home `~/.gitconfig` on your personal machine might look like:

    [user]
      email = "john@home.org"

Whereas at work it might be:

    [user]
      email = "john@company.com"

To handle this, on each machine create a configuration file called
`~/.config/chezmoi/chezmoi.toml` defining what might change. For your home
machine:

    [data]
      email = "john@home.org"

If you intend to store private data (e.g. access tokens) in
`~/.config/chezmoi/chezmoi.toml`, make sure it has permissions `0600`. See
"Keeping data private" below for more discussion on this.

If you prefer, you can use any format supported by
[Viper](https://github.com/spf13/viper) for your configuration file. This
includes JSON, YAML, and TOML.

Then, add `~/.gitconfig` to `chezmoi` using the `-T` flag to automatically turn
it in to a template:

    chezmoi add -T ~/.gitconfig

You can then open the template (which will be saved in the file
`~/.local/share/chezmoi/dot_gitconfig.tmpl`):

    chezmoi edit ~/.gitconfig

The file should look something like:

    [user]
      email = "{{ .email }}"

`chezmoi` will substitute the variables from the `data` section of your
`~/.config/chezmoi/chezmoi.toml` file when calculating the target state of
`.gitconfig`.

For more advanced usage, you can use the full power of the
[`text/template`](https://godoc.org/text/template) language to include or
exclude sections of file. `chezmoi` provides the following automatically
populated variables:

| Variable                | Value                                                                                                                  |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `.chezmoi.arch`         | Architecture, e.g. `amd64`, `arm`, etc. as returned by [runtime.GOARCH](https://godoc.org/runtime#pkg-constants).      |
| `.chezmoi.fullHostname` | The full hostname of the machine `chezmoi` is running on.                                                              |
| `.chezmoi.group`        | The group of the user running `chezmoi`.                                                                               |
| `.chezmoi.homedir`      | The home directory of the user running `chezmoi`.                                                                      |
| `.chezmoi.hostname`     | The hostname of the machine `chezmoi` is running on, up to the first `.`.                                              |
| `.chezmoi.os`           | Operating system, e.g. `darwin`, `linux`, etc. as returned by [runtime.GOOS](https://godoc.org/runtime#pkg-constants). |
| `.chezmoi.osRelease`    | The information from `/etc/os-release`, Linux only, run `chezmoi data` to see its output.                              |
| `.chezmoi.username`     | The username of the user running `chezmoi`.                                                                            |

For a full list of variables, run:

    chezmoi data

For example, in your `~/.local/share/chezmoi/dot_bashrc.tmpl` you might have:

    # common config
    export EDITOR=vi

    # machine-specific configuration
    {{- if eq .chezmoi.hostname "work-laptop" }}
    # this will only be included in ~/.bashrc on work-laptop
    {{- end }}

`chezmoi` includes all of the hermetic text functions from
[`sprig`](http://masterminds.github.io/sprig/).

If, after executing the template, the file contents are empty, the target file
will be removed. This can be used to ensure that files are only present on
certain machines. If you want an empty file to be created anyway, you will need
to give it an `empty_` prefix. See "Under the hood" below.

For coarser-grained control of files and entire directories are managed on
different machines, or to exclude certain files completely, you can create
`.chezmoiignore` files in the source directory. These specify a list of patterns
that `chezmoi` should ignore, and are interpreted as templates. An example
`.chezmoiignore` file might look like:

    README.md
    {{- if ne .chezmoi.hostname "work-laptop" }}
    .work # only manage .work on work-laptop
    {{- end }}

## Keeping data private

`chezmoi` automatically detects when files and directories are private when
adding them by inspecting their permissions. Private files and directories are
stored in `~/.local/share/chezmoi` as regular, public files with permissions
`0644` and the name prefix `private_`. For example:

    chezmoi add ~/.netrc

will create `~/.local/share/chezmoi/private_dot_netrc` (assuming `~/.netrc` is
not world- or group- readable, as it should be). This file is still private
because `~/.local/share/chezmoi` is not group- or world- readable or executable.
`chezmoi` checks that the permissions of `~/.local/share/chezmoi` are `0700` on
every run and will print a warning if they are not.

It is common that you need to store access tokens in config files, e.g. a
[Github access
token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).
There are several ways to keep these tokens secure, and to prevent them leaving
your machine.

### Using templates variables

Typically, `~/.config/chezmoi/chezmoi.toml` is not checked in to version control
and has permissions 0600. You can store tokens as template values in the `data`
section. For example, if your `~/.config/chezmoi/chezmoi.toml` contains:

    [data]
      [data.github]
        user = "<github-username>"
        token = "<github-token>"

Your `~/.local/share/chezmoi/private_dot_gitconfig.tmpl` can then contain:

    {{- if (index . "github") }}
    [github]
      user = "{{ .github.user }}"
      token = "{{ .github.token }}"
    {{- end }}

Any config files containing tokens in plain text should be private (permissions
`0600`).

### Using 1Password

`chezmoi` includes support for [1Password](https://1password.com/) using the
[1Password CLI](https://support.1password.com/command-line-getting-started/) to
expose data as a template function.

Log in and get a session using:

    eval $(op login <subdomain>.1password.com <email>)

The structured data from `op get item <uuid>` is available as the `onepassword`
template function, for example:

    {{ (onepassword "<uuid>").details.password }}

### Using Bitwarden

`chezmoi` includes support for [Bitwarden](https://bitwarden.com/) using the
[Bitwarden CLI](https://github.com/bitwarden/cli) to expose data as a template
function.

Log in to Bitwarden using:

    bw login <bitwarden-email>

Unlock your Bitwarden vault:

    bw unlock

Set the `BW_SESSION` environment variable, as instructed.

The structured data from `bw get` is available as the `bitwarden` template
function in your config files, for example:

    username = {{ (bitwarden "item" "example.com").login.username }}
    password = {{ (bitwarden "item" "example.com").login.password }}

### Using LastPass

`chezmoi` includes support for [LastPass](https://lastpass.com) using the
[LastPass CLI](https://lastpass.github.io/lastpass-cli/lpass.1.html) to expose
data as a template function.

Log in to LastPass using:

    lpass login <lastpass-username>

Check that `lpass` is working correctly by showing password data:

    lpass show --json <lastpass-entry-id>

where `<lastpass-entry-id>` is a [LastPass Entry
Specification](https://lastpass.github.io/lastpass-cli/lpass.1.html#_entry_specification).

The structured data from `lpass show --json id` is available as the `lastpass`
template function. The value will be an array of objects. You can use the
`index` function and `.Field` syntax of the `text/template` language to extract
the field you want. For example, to extract the `password` field from first the
"Github" entry, use:

    githubPassword = "{{ (index (lastpass "Github") 0).password }}"

`chezmoi` automatically parses the `note` value of the Lastpass entry, so, for
example, you can extract a private SSH key like this:

    {{ (index (lastpass "SSH") 0).note.privateKey }}

Keys in the `note` section written as `CamelCase Words` are converted to
`camelCaseWords`.

### Using pass

`chezmoi` includes support for [pass](https://www.passwordstore.org/) using the
`pass` CLI.

The first line of the output of `pass show <pass-name>` is available as the
`pass` template function, for example:

    {{ pass "<pass-name>" }}

### Using Vault

`chezmoi` includes support for [Vault](https://www.vaultproject.io/) using the
[Vault CLI](https://www.vaultproject.io/docs/commands/) to expose data as a
template function.

The vault CLI needs to be correctly configured on your machine, e.g. the
`VAULT_ADDR` and `VAULT_TOKEN` environment variables must be set correctly.
Verify that this is the case by running:

    vault kv get -format=json <key>

The stuctured data from `vault kv get -format=json` is available as the `vault`
template function. You can use the `.Field` syntax of the `text/template`
language to extract the data you want. For example:

    {{ (vault "<key>").data.data.password }}

### Using keyring

`chezmoi` includes support for Keychain (on macOS), GNOME Keyring (on Linux),
and Windows Credentials Manager (on Windows) via the
[`zalando/go-keyring`](https://github.com/zalando/go-keyring) library.

Set passwords with:

    $ chezmoi keyring set --service=<service> --user=<user>
    Password: xxxxxxxx

The password can then be used in templates using the `keyring` function which
takes the service and user as arguments.

For example, save a Github access token in keyring with:

    $ chezmoi keyring set --service=github --user=<github-username>
    Password: xxxxxxxx

and then include it in your `~/.gitconfig` file with:

    [github]
      user = "{{ .github.user }}"
      token = "{{ keyring "github" .github.user }}"

You can query the keyring from the command line:

    chezmoi keyring get --service=github --user=<github-username>

### Using a generic secret manager

You can use any command line tool that outputs secrets either as a string or in
JSON format. Choose the binary by setting `genericSecret.command` in your
configuration file. You can then invoke this command with the `secret` and
`secretJSON` template functions which return the raw output and JSON-decoded
output respectively. All of the above secret managers can be supported in this
way:

| Secret Manager  | `genericSecret.command` | Template skeleton                                 |
| --------------- | ----------------------- | ------------------------------------------------- |
| 1Password       | `op`                    | `{{ secretJSON "get" "item" <id> }}`              |
| Bitwarden       | `bw`                    | `{{ secretJSON "get" <id> }}`                     |
| Hashicorp Vault | `vault`                 | `{{ secretJSON "kv" "get" "-format=json" <id> }}` |
| LastPass        | `lpass`                 | `{{ secretJSON "show" "--json" <id> }}`           |
| pass            | `pass`                  | `{{ secret "show" <id> }}`                        |

### Encrypting individual files with `gpg` (beta)

`chezmoi` supports encrypting individual files with
[`gpg`](https://www.gnupg.org/). Specify the encryption key to use in your
configuration file (`chezmoi.toml`) with the `gpgReceipient` key:

    gpgRecipient = "..."

Add files to be encrypted with the `--encrypt` flag, for example:

    chezmoi add --encrypt ~/.ssh/id_rsa

`chezmoi` will encrypt the file with

    gpg --armor --encrypt --recipient $gpgRecipient

and store the encrypted file in the source state. The file will automatically be
decrypted when generating the target state.

This feature is still in beta and has a couple of rough edges:

* Editing an encrypted file will edit the cyphertext, not the plaintext.
* Diff'ing an encrypted file will show the difference between the old plaintext
  and the new cyphertext.

### Using encrypted config files

`chezmoi` takes a `-c` flag specifying the file to read its configuration from.
You can encrypt your configuration and then only decrypt it when needed:

    gpg -d ~/.config/chezmoi/chezmoi.toml.gpg | chezmoi -c /dev/stdin apply

## Importing archives

It is occasionally useful to import entire archives of configuration into your
source state. The `import` command does this. For example, to import the latest
version
[`github.com/robbyrussell/oh-my-zsh`](https://github.com/robbyrussell/oh-my-zsh)
to `~/.oh-my-zsh` run:

    curl -s -L -o oh-my-zsh-master.tar.gz https://github.com/robbyrussell/oh-my-zsh/archive/master.tar.gz
    chezmoi import --strip-components 1 --destination ~/.oh-my-zsh oh-my-zsh-master.tar.gz

Note that this only updates the source state. You will need to run

    chezmoi apply

to update your destination directory.

## Exporting archives

`chezmoi` can create an archive containing the target state. This can be useful
for generating target state on a different machine or for simply inspecting the
target state. A particularly useful command is:

    chezmoi archive | tar tvf -

which lists all the targets in the target state.

## Using non-`git` version control systems

By default, `chezmoi` uses `git`, but you can use any version control system of
your choice. In your config file, specify the command to use. For example, to
use Mercurial specify:

    [sourceVCS]
      command = "hg"

The source VCS command is used in the `chezmoi` commands `init`, `source`, and
`update`, and support for VCSes other than `git` is limited but easy to add. If
you'd like to see your VCS better supported, please [open an issue on
Github](https://github.com/twpayne/chezmoi/issues/new).

## Under the hood

For an example of how `chezmoi` stores its state, see
[`github.com/twpayne/dotfiles`](https://github.com/twpayne/dotfiles).

`chezmoi` stores the desired state of files, symbolic links, and directories in
regular files and directories in `~/.local/share/chezmoi`. This location can be
overridden with the `-S` flag or by giving a value for `sourceDir` in
`~/.config/chezmoi/chezmoi.toml`.  Some state is encoded in the source names.
`chezmoi` ignores all files and directories in the source directory that begin
with a `.`. The following prefixes and suffixes are special, and are
collectively referred to as "attributes":

| Prefix/suffix        | Effect                                                                            |
| -------------------- | ----------------------------------------------------------------------------------|
| `encrypted_` prefix  | Encrypt the file in the source state.                                             |
| `private_` prefix    | Remove all group and world permissions from the target file or directory.         |
| `empty_` prefix      | Ensure the file exists, even if is empty. By default, empty files are removed.    |
| `exact_` prefix      | Remove anything not managed by `chezmoi`.                                         |
| `executable_` prefix | Add executable permissions to the target file.                                    |
| `symlink_` prefix    | Create a symlink instead of a regular file.                                       |
| `dot_` prefix        | Rename to use a leading dot, e.g. `dot_foo` becomes `.foo`.                       |
| `.tmpl` suffix       | Treat the contents of the source file as a template.                              |

Order is important, the order is `exact_`, `private_`, `empty_`, `executable_`,
`symlink_`, `dot_`, `.tmpl`.

Different target types allow different prefixes and suffixes:

| Target type   | Allowed prefixes and suffixes                                      |
| ------------- | ------------------------------------------------------------------ |
| Directory     | `exact_`, `private_`, `dot_`                                       |
| Regular file  | `encrypted_`, `private_`, `empty_`, `executable_`, `dot_`, `.tmpl` |
| Symbolic link | `symlink_`, `dot_`, `.tmpl`                                        |

You can change the attributes of a target in the source state with the `chattr`
command. For example, to make `~/.netrc` private and a template:

    chezmoi chattr private,template ~/.netrc

This only updates the source state of `~/.netrc`, you will need to run `apply`
to apply the changes to the destination state:

    chezmoi apply ~/.netrc

## Using `chezmoi` outside your home directory

`chezmoi`, by default, operates on your home directory, but this can be
overridden with the `--dest` command line flag or by specifying `destDir` in
your `~/.config/chezmoi/chezmoi.toml`. In theory, you could use `chezmoi` to
manage any aspect of your filesystem. That said, although you can do this, you
probably shouldn't. Existing configuration management tools like
[Puppet](https://puppet.com/), [Chef](https://www.chef.io/chef/),
[Ansible](https://www.ansible.com/), and [Salt](https://www.saltstack.com/) are
much better suited to whole system configuration management.

`chezmoi` was inspired by Puppet, but created because Puppet is a slow overkill
for managing your personal configuration files. The focus of `chezmoi` will
always be personal home directory management. If your needs grow beyond that,
switch to a whole system configuration management tool.

## `chezmoi` in the news

* [Linux Fu: The Kitchen Sink on
  hackaday.com](https://hackaday.com/2019/01/10/linux-fu-the-kitchen-sync/)

* [chezmoi on
  reddit.com/r/linux](https://www.reddit.com/r/linux/comments/afogsb/chezmoi_manage_your_dotfiles_securely_across/)

* [chezmoi on lobste.rs](https://lobste.rs/stories/uet36y/)

* [chezmoi on
  news.ycombinator.com](https://news.ycombinator.com/item?id=18902090)

## Related projects

See [`dotfiles.github.io`](https://dotfiles.github.io/).

## License

MIT
