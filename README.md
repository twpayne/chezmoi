# `chezmoi`

[![Build Status](https://travis-ci.org/twpayne/chezmoi.svg?branch=master)](https://travis-ci.org/twpayne/chezmoi)
[![GoDoc](https://godoc.org/github.com/twpayne/chezmoi?status.svg)](https://godoc.org/github.com/twpayne/chezmoi)
[![Report Card](https://goreportcard.com/badge/github.com/twpayne/chezmoi)](https://goreportcard.com/report/github.com/twpayne/chezmoi)

Manage your dotfiles securely across multiple machines.


## Features

 * Declarative: you declare the desired state of files, directories, and
   symbolic links in your home directory and `chezmoi` updates your home
directory to match that state.

 * Flexible: your dotfiles can be templates (using
   [`text/template`](https://godoc.org/text/template) syntax). Predefined
variables allow you to change behaviour depending on operating system,
architecture, and hostname.

 * Secure: `chezmoi` can retreive secrets from
   [Bitwarden](https://bitwarden.com/), [LastPass](https://lastpass.com/), your
Keychain (on macOS), and [GNOME
Keyring](https://wiki.gnome.org/Projects/GnomeKeyring) (on Linux).

 * Robust: `chezmoi` updates all files and symbolic links atomically (using
   [`google/renameio`](https://github.com/google/renameio)) so you are never
left with incomplete files that could lock you out, even if the update process
is interrupted.

 * Portable: `chezmoi`'s configuration uses only visible, regular files and
   directories and so is portable across version control systems and operating
systems.

 * Transparent: `chezmoi` includes verbose and dry run modes so you can review
   exactly what changes it will make to your home directory before making them.

 * Fast, easy to use, and familiar: `chezmoi` runs in fractions of a second and
   includes commands to make most operations trivial. You can use the version
control system of your choice to manage your configuration, and many different
formats (e.g. JSON, YAML, TOML, etc.) are supported for the configuration file.


## I already have a system to manage my dotfiles, why should I use `chezmoi`?

 * If your system is based on copying files with a shell script or creating
   symlinks (e.g. using [GNU
Stow](http://brandon.invergo.net/news/2012-05-26-using-gnu-stow-to-manage-your-dotfiles.html))
then handling files that vary from machine to machine requires manual work. You
might need to maintain separate config files for separate machines, or run
different commands on different machines. `chezmoi` gives you a single command
that works on every machine.

 * If your system is based on using `git` with a different branches for
   different machines, then you need manually merge or rebase to ensure that
changes you make are applied to each machine. `chezmoi` makes it trivial to
share common parts while allowing specific per-machine configuration.

 * If your system stores secrets in plain text, then your dotfiles repository
   must be private. With `chezmoi` you never need to store secrets in your
repository, so you can make it public. You can check out your repository on
your work machine and not fear that this will give your work IT department to
access to your personal data.

 * If your system was written by you for your personal use, then it probably
   has the minimum functionality that you need. `chezmoi` includes a wide range
of functionality out-of-the-box, including dry run and diff modes.

 * All systems suffer from the "bootstrap" problem: you need to install your
   system before you can install your dotfiles. `chezmoi` provides
statically-linked binaries, packages for many Linux and BSD distributions, and
Homebrew formulae to make overcoming the bootstrap problem as simple as possible.


## Installation

Binaries and packages for multiple platforms, including RedHat, Debian,
FreeBSD, and OpenBSD, are available on the [releases
page](https://github.com/twpayne/chezmoi/releases).

On macOS you can install `chezmoi` with Homebrew:

    $ brew install twpayne/taps/chezmoi

If you have Go installed you can install the latest version from `HEAD`:

    $ go get -u github.com/twpayne/chezmoi


## Quick start

`chezmoi` evaluates the source state for the current machine and then updates
the destination directory, where:

 * The *source state* declares the desired state of your home directory,
   including templates and machine-specific configuration.

 * The *source directory* is where `chezmoi` stores the source state, by
   default `~/.config/share/chezmoi`.

 * The *target state* is the source state computed for the current machine.

 * The *destination directory* is the directory that `chezmoi` manages, by
   default `~`, your home directory.

 * A *target* is a file, directory, or symlink in the destination directory.

 * The *destination state* is the state of all the targets in the destination
   directory.

 * The *config file* contains machine-specific configuration, by default it is
   `~/.config/chezmoi/chezmoi.yaml`.

Manage an existing file with `chezmoi`:

    $ chezmoi add ~/.bashrc

This will create the source directory `~/.local/share/chezmoi` with permissions
`0600` where `chezmoi` will store the source state (if it does not already
exist), and copy `~/.bashrc` to `~/.local/share/chezmoi/dot_bashrc`.

You should manage your `~/.local/share/chezmoi` directory with the version
control system of your choice. `chezmoi` will ignore all files and directories
beginning with a `.` in this directory, including directories like `.git` and
`.hg`.

Edit the source state:

    $ chezmoi edit ~/.bashrc

This will open `~/.local/share/chezmoi/dot_bashrc` in your `$EDITOR`. Make some
changes and save them.

See what changes `chezmoi` would make:

    $ chezmoi diff

Apply the changes:

    $ chezmoi -v apply

All `chezmoi` commands accept the `-v` (verbose) flag to print out exactly what
changes they will make to the file system, and the `-n` (dry run) flag to not
make any actual changes. The combination `-n` `-v` is very useful if you want
to see exactly what changes would be made.

For a full list of commands run:

    $ chezmoi help


## Using templates to manage files that vary from machine to machine

The primary goal of `chezmoi` is to manage configuration files across multiple
machines, for example your personal macOS laptop, your work Ubuntu desktop, and
your work Linux laptop. You will want to keep much configuration the same
across these, but also need machine-specific configurations for email
addresses, credentials, etc. `chezmoi` achieves this functionality by using
[`text/template`](https://godoc.org/text/template) for the source state where
needed.

For example, your home `~/.gitconfig` on your personal machine might look like:

    [user]
        email = john@home.org

Whereas at work it might be:

    [user]
        email = john@company.com

To handle this, on each machine create a configuration file called
`~/.config/chezmoi/chezmoi.yaml` defining what might change. For your home
machine:

    data:
      email: john@home.org

If you intend to store private data (e.g. access tokens) in
`~/.config/chezmoi/chezmoi.yaml`, make sure it has permissions `0600`. See
"Keeping data private" below for more discussion on this.

If you prefer, you can use any format supported by
[Viper](https://github.com/spf13/viper) for your configuration file. This
includes JSON, YAML, and TOML.

Then, add `~/.gitconfig` to `chezmoi` using the `-T` flag to automatically turn
it in to a template:

    $ chezmoi add -T ~/.gitconfig

You can then open the template (which will be saved in the file
`~/.local/share/chezmoi/dot_gitconfig.tmpl`):

    $ chezmoi edit ~/.gitconfig

The file should look something like:

    [user]
        email = {{ .email }}

`chezmoi` will substitute the variables from the `data` section of your
`~/.config/chezmoi/chezmoi.yaml` file when calculating the target state of
`.gitconfig`.

For more advanced usage, you can use the full power of the
[`text/template`](https://godoc.org/text/template) language to include or
exclude sections of file. `chezmoi` provides the following automatically
populated variables:

| Variable            | Value                                                                                                                  |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `.chezmoi.arch`     | Architecture, e.g. `amd64`, `arm`, etc. as returned by [runtime.GOARCH](https://godoc.org/runtime#pkg-constants).      |
| `.chezmoi.group`    | The group of the user running `chezmoi`.                                                                               |
| `.chezmoi.homedir`  | The home directory of the user running `chezmoi`.                                                                      |
| `.chezmoi.hostname` | The hostname of the machine `chezmoi` is running on.                                                                   |
| `.chezmoi.os`       | Operating system, e.g. `darwin`, `linux`, etc. as returned by [runtime.GOOS](https://godoc.org/runtime#pkg-constants). |
| `.chezmoi.username` | The username of the user running `chezmoi`.                                                                            |

For example, in your `~/.local/share/chezmoi/dot_bashrc.tmpl` you might have:

    # common config
    export EDITOR=vi

    # machine-specific configuration
    {{- if eq .chezmoi.hostname "work-laptop" }}
    # this will only be included in ~/.bashrc on work-laptop
    {{- end }}

If, after executing the template, the file contents are empty, the target file
will be removed. This can be used to ensure that files are only present on
certain machines. If you want an empty file to be created anyway, you will need
to give it an `empty_` prefix. See "Under the hood" below.

For coarser-grained control of files and entire directories are managed on
different machines, or to exclude certain files completely, you can create
`.chezmoiignore` files in the source directory. These specify a list of
patterns that `chezmoi` should ignore, and are interpreted as templates. An
example `.chezmoiignore` file might look like:

    README.md
    {{- if ne .chezmoi.hostname "work-laptop" }}
    .work # only manage .work on work-laptop
    {{- end }}


## Keeping data private

`chezmoi` automatically detects when files and directories are private when
adding them by inspecting their permissions. Private files and directories are
stored in `~/.local/share/chezmoi` as regular, public files with permissions
`0644` and the name prefix `private_`. For example:

    $ chezmoi add ~/.netrc

will create `~/.local/share/chezmoi/private_dot_netrc` (assuming `~/.netrc` is
not world- or group- readable, as it should be). This file is still private
because `~/.local/share/chezmoi` is not group- or world- readable or
executable.  `chezmoi` checks that the permissions of `~/.local/share/chezmoi`
are `0700` on every run and will print a warning if they are not.

It is common that you need to store access tokens in config files, e.g. a
[Github access
token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).
There are several ways to keep these tokens secure, and to prevent them leaving
your machine.

### Using templates variables

Typically, `~/.config/chezmoi/chezmoi.yaml` is not checked in to version
control and has permissions 0600. You can store tokens as template values in
the `data` section. For example, if your `~/.config/chezmoi/chezmoi.yaml`
contains:

    data:
      github:
        user: <github-username>
        token: <github-token>

Your `~/.local/share/chezmoi/private_dot_gitconfig.tmpl` can then contain:

    {{- if .github }}
    [github]
        user = {{ .github.user }}
        token = {{ .github.token }}
    {{- end }}

Any config files containing tokens in plain text should be private (permissions
`0600`).

### Using Bitwarden

`chezmoi` includes support for [Bitwarden](https://bitwarden.com/) using the
[Bitwarden CLI](https://github.com/bitwarden/cli) to expose data as a template
function.

Log in to Bitwarden using:

    $ bw login <bitwarden-email>

Unlock your Bitwarden vault:

    $ bw unlock

Set the `BW_SESSION` environment variable, as instructed. You can also pass the
session directly to `chezmoi` using the `--bitwarden-session` flag.

The structured data from `bw get` is available as the `bitwarden` template
function in your config files, for example:

    username = {{ (bitwarden "item" "example.com").login.username }}
    password = {{ (bitwarden "item" "example.com").login.password }}

### Using LastPass

`chezmoi` includes support for [LastPass](https://lastpass.com) using the
[LastPass CLI](https://lastpass.github.io/lastpass-cli/lpass.1.html) to expose
data as a template function.

Log in to LastPass using:

    $ lpass login <lastpass-username>

Check that `lpass` is working correctly by showing password data:

    $ lpass show -j <lastpass-entry-id>

where `<lastpass-entry-id>` is a [LastPass Entry
Specification](https://lastpass.github.io/lastpass-cli/lpass.1.html#_entry_specification).

The structured data from `lpass show -j id` is available as the `lastpass`
template function. The value will be an array of objects. You can use the
`index` function and `.Field` syntax of the `text/template` language to extract
the field you want. For example, to extract the `password` field from first the
"Github" entry, use:

    githubPassword = {{ (index (lastpass "Github") 0).password }}

`chezmoi` automatically parses the `note` value of the Lastpass entry, so, for
example, you can extract a private SSH key like this:

    {{ (index (lastpass "SSH") 0).note.privateKey }}

Keys in the `note` section written as `CamelCase Words` are converted to
`camelCaseWords`.

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
        user = {{ .github.user }}
        token = {{ keyring "github" .github.user }}

You can query the keyring from the command line:

    $ chezmoi keyring get --service=github --user=<github-username>

### Using encrypted config files

`chezmoi` takes a `-c` flag specifying the file to read its configuration from.
You can encrypt your configuration and then only decrypt it when needed:

    $ gpg -d ~/.config/chezmoi/chezmoi.yaml.gpg | chezmoi -c /dev/stdin apply


## Managing your `~/.chezmoi` directory with version control

`chezmoi` has some helper commands to assist managing your source directory
with version control. The default version control system is `git` but you can
change this by setting `sourceVCSCommand` in your
`~/.config/chezmoi/chezmoi.yaml` file, for example, if you want to use
Mercurial:

    sourceVCSCommand: hg

`chezmoi source` is then a shortcut to running `sourceVCSCommand` in your
`~/.local/share/chezmoi` directory. For example you can push the current branch
with:

    $ chezmoi source push

Extra arguments are passed along unchanged, although you'll need to use `--`
stop `chezmoi` from interpreting extra flags. For example:

    $ chezmoi source pull -- --rebase

The `source` command accepts the usual `-n` and `-v` flags, so you can see
exactly what it will run without executing it.

As a shortcut,

    $ chezmoi cd

starts a shell in your source directory, which can be very useful when
performing multiple VCS operations.


## Importing archives

It is occasionally useful to import entire archives of configuration into your
source state. The `import` command does this. For example, to import the
latest version
[`github.com/robbyrussell/oh-my-zsh`](https://github.com/robbyrussell/oh-my-zsh)
to `~/.oh-my-zsh` run:

    $ curl -s -L -o oh-my-zsh-master.tar.gz https://github.com/robbyrussell/oh-my-zsh/archive/master.tar.gz
    $ chezmoi import --strip-components 1 --destination ~/.oh-my-zsh oh-my-zsh-master.tar.gz

Note that this only updates the source state. You will need to run

    $ chezmoi apply

to update your destination directory.


## Exporting archives

`chezmoi` can create an archive containing the target state. This can be useful
for generating target state on a different machine or for simply inspecting the
target state. A particularly useful command is:

    $ chezmoi archive | tar tvf -

which lists all the targets in the target state.


## Under the hood

For an example of how `chezmoi` stores its state, see
[`github.com/twpayne/dotfiles`](https://github.com/twpayne/dotfiles).

`chezmoi` stores the desired state of files, symbolic links, and directories in
regular files and directories in `~/.local/share/chezmoi`. This location can be
overridden with the `-S` flag or by giving a value for `sourceDir` in
`~/.config/chezmoi/chezmoi.yaml`.  Some state is encoded in the source names.
`chezmoi` ignores all files and directories in the source directory that begin
with a `.`. The following prefixes and suffixes are special, and are
collectively referred to as "attributes":

| Prefix/suffix        | Effect                                                                            |
| -------------------- | ----------------------------------------------------------------------------------|
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

| Target type   | Allowed prefixes and suffixes                        |
| ------------- | ---------------------------------------------------- |
| Directory     | `exact_`, `private_`, `dot_`                         |
| Regular file  | `private_`, `empty_`, `executable_`, `dot_`, `.tmpl` |
| Symbolic link | `symlink_`, `dot_`, `.tmpl`                          |

You can change the attributes of a target in the source state with the `chattr`
command. For example, to make `~/.netrc` private and a template:

    chezmoi chattr private,template ~/.netrc

This only updates the source state of `~/.netrc`, you will need to run `apply`
to apply the changes to the destination state:

    chezmoi apply ~/.netrc


## Using `chezmoi` outside your home directory

`chezmoi`, by default, operates on your home directory, but this can be
overridden with the `--dest` command line flag or by specifying `destDir` in
your `~/.config/chezmoi/chezmoi.yaml`. In theory, you could use `chezmoi` to
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

 * [Linux Fu: The Kitchen Sink on hackaday.com](https://hackaday.com/2019/01/10/linux-fu-the-kitchen-sync/)
 * [chezmoi: manage your dotfiles securely across multiple machines on reddit.com/r/linux](https://www.reddit.com/r/linux/comments/afogsb/chezmoi_manage_your_dotfiles_securely_across/)


## License

The MIT License (MIT)

Copyright (c) 2018 Tom Payne

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
