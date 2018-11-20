# `chezmoi`

[![Build Status](https://travis-ci.org/twpayne/chezmoi.svg?branch=master)](https://travis-ci.org/twpayne/chezmoi)
[![GoDoc](https://godoc.org/github.com/twpayne/chezmoi?status.svg)](https://godoc.org/github.com/twpayne/chezmoi)
[![Report Card](https://goreportcard.com/badge/github.com/twpayne/chezmoi)](https://goreportcard.com/report/github.com/twpayne/chezmoi)

`chezmoi` is a tool for managing your home directory across multiple machines.


## Features

 * Declarative: you declare the desired state of files and directories in your
   home directory and `chezmoi` updates your home directory to match that
state.

 * Flexible: your dotfiles can be templates (using
   [text/template](https://godoc.org/text/template) syntax). Predefined
variables allow you to change behaviour depending on operating system,
architecture, and hostname.

 * Robust: `chezmoi` updates all files atomically (using
   [google/renameio](https://github.com/google/renameio)) so you are never left
with incomplete files that could lock you out, even if the update process is
interrupted.

 * Portable: `chezmoi`'s configuration uses only visible, regular files and
   directories and so is portable across version control systems and operating
systems.

 * Transparent: `chezmoi` includes verbose and dry run modes so you can review
   exactly what changes it will make to your home directory before making them.

 * Practical: `chezmoi` manages hidden files (dot files), directories, private,
   and executable files.

 * Fast, easy to use, and familiar: `chezmoi` runs in fractions of a second and
   includes commands to make most operations trivial. You can use the version
control system of your choice to manage your configuration.


## Quick start

Install it:

    $ go install github.com/twpayne/chezmoi

Start using it:

    $ chezmoi init

This creates a directory called `~/.chezmoi` with permissions 0600 where
`chezmoi` will store its state. You should manage this directory with the
version control system of your choice. `chezmoi` will ignore all files and
directories beginning with a `.` in this directory.

Manage an existing file with `chezmoi`:

    $ chezmoi add .bashrc

This copies `~/.bashrc` to `~/.chezmoi/dot_bashrc`.

Edit the desired state:

    $ chezmoi edit .bashrc

This will open `~/.chezmoi/dot_bashrc` in your `$EDITOR`. Make some changes and
save them.

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
[text/template](https://godoc.org/text/template) for the source configuration
files where needed.

For example, your home `~/.gitconfig` on your personal machine might look like:

    [user]
        name = John Smith
        email = john@home.org

Whereas at work it might be:

    [user]
        name = John Smith
        email = john@company.com

To handle this, on each machine create a file called `~/.chezmoi.yaml` defining
what might change. For your home machine:

    data:
      name: John Smith
      email: john@home.org

If you intend to store private data (e.g. access tokens) in `~/.chezmoi.yaml`,
make sure it has mode 0600. See "Keeping data private" below for more
discussion on this.

Then, add `~/.gitconfig` to `chezmoi` using the `-T` flag to automatically turn
it in to a template:

    $ chezmoi add -T .gitconfig

You can then open the template (which will be saved in the file
`~/.chezmoi/dot_gitconfig.tmpl`):

    $ chezmoi edit .gitconfig

The file should look something like:

    [user]
        name = {{ .name }}
        email = {{ .email }}

`chezmoi` will substitute the variables from the `data` section of your
`~/.chezmoi.yaml` file when calculating the desired state of `.gitconfig`.

For more advanced usage, you can use the full power of the
[text/template](https://godoc.org/text/template) language to include or exclude
sections of file. `chezmoi` provides the following automatically populated
variables:

| Variable           | Value                                                                                                                  |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| `chezmoi.arch`     | Architecture, e.g. `amd64`, `arm`, etc. as returned by [runtime.GOARCH](https://godoc.org/runtime#pkg-constants).      |
| `chezmoi.group`    | The group of the user running `chezmoi`.                                                                               |
| `chezmoi.homedir`  | The home directory of the user running `chezmoi`.                                                                      |
| `chezmoi.hostname` | The hostname of the machine `chezmoi` is running on.                                                                   |
| `chezmoi.os`       | Operating system, e.g. `darwin`, `linux`, etc. as returned by [runtime.GOOS](https://godoc.org/runtime#pkg-constants). |
| `chezmoi.username` | The username of the user running `chezmoi`.                                                                            |

For example, in your `~/.chezmoi/dot_bashrc.tmpl` you might have:

    # common config
    export EDITOR=vi

    # machine-specific configuration
    {{- if .chezmoi.hostname eq "work-laptop" }}
    # this will only be included in ~/.bashrc on work-laptop
    {{- end }}


## Keeping data private

`chezmoi` automatically detects when files and directories are private when
adding them by inspecting their permissions. Private files and directories are
stored in `~/.chezmoi` as regular, public files with permissions 0644 and the
name prefix `private_`. For example:

    $ chezmoi add .netrc

will create `~/.chezmoi/private_dot_netrc`. Note that this file is still
private because `~/.chezmoi` is not group- or world- readable or executable.
`chezmoi` checks that the permissions of `~/.chezmoi` are 0700 on every run and
will print a warning if they are not.

It is common that you need to store access tokens in config files, e.g. a
[Github access
token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).
There are several ways to keep these tokens secure, and to prevent them leaving
your machine.

### Using templates

Typically, `~/.chezmoi.yaml` is not checked in to version control and has
permissions 0600. You can store tokens as template values in the `data`
section. For example, mine has:

    data:
      github:
        user: twpayne
        token: xxxxxxxx

My `~/.chezmoi/private_dot_gitconfig.tmpl` then contains:

    {{- if .github }}
    [github]
        user = {{ .github.user }}
        token = {{ .github.token }}
    {{- end }}

Note that any config files containing tokens in plain text should be private
(mode 0600).

### Using encrypted config files

`chezmoi` takes a `-c` flag specifying the file to read its configuration from.
You can encrypt your configuration and then only decrypt it when needed:

    $ gpg -d ~/.chezmoi.yaml.gpg | chezmoi -c /dev/stdin apply


## Managing your `~/.chezmoi` directory with version control

`chezmoi` has some helper commands to assist managing your source directory
with version control. The default version control system is `git` but you can
change this by setting `sourceVCSCommand` in your `.chezmoi.yaml` file, for
example, if you want to use Mercurial:

    sourceVCSCommand: hg

`chezmoi source` is then a shortcut to running `sourceVCSCommand` in your
`~/.chezmoi` directory. For example you can push the current branch with:

    $ chezmoi source push

Extra arguments are passed along unchanged, although you'll need to use `--`
stop `chezmoi` from interpreting extra flags. For example:

    $ chezmoi source pull -- --rebase

The `source` command accepts the usual `-n` and `-v` flags, so you can see
exactly what it will run without executing it.


## Under the hood

`chezmoi` stores the desired state of files and directories in `~/.chezmoi`.
This location can be overridden with the `-s` flag or by giving a value for
`sourceDir` in `~/.chezmoi.yaml`. Some state is encoded in the source names.
`chezmoi` ignores all files and directories in the source directory that begin
with a `.`.  The following prefixes and suffixes are special.

| Prefix               | Effect                                                                            |
| -------------------- | ----------------------------------------------------------------------------------|
| `private_` prefix    | Remove all group and world permissions from the target file or directory.         |
| `executable_` prefix | Add executable permissions to the target file.                                    |
| `dot_` prefix        | Rename the file or directory to use a leading dot, e.g. `dot_foo` becomes `.foo`. |
| `.tmpl` suffix       | Treat the source file as a template.                                              |

Order is important, the order is `private_`, `executable_`, `dot_`, `.tmpl`.


## Using `chezmoi` outside your home directory

`chezmoi`, by default, operates on your home directory, but this can be
overridden with the `--target` command line flag or by specifying `targetDir`
in your `~/.chezmoi.yaml`. In theory, you could use `chezmoi` to manage any
aspect of your filesystem. That said, although you can do this, you probably
shouldn't. Existing configuration management tools like
[Puppet](https://puppet.com/), [Chef](https://www.chef.io/chef/),
[Ansible](https://www.ansible.com/), and [Salt](https://www.saltstack.com/) are
much better suited to whole system configuration management.

`chezmoi` was inspired by Puppet, but created because Puppet is a slow overkill
for managing your personal configuration files. The focus of `chezmoi` will
always be personal home directory management. If your needs grow beyond that,
switch to a whole system configuration management tool.
