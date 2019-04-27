# Getting started with `chezmoi`

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
