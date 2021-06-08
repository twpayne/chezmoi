# chezmoi changes

<!--- toc --->
* [Overview](#overview)
* [Version 2](#version-2)
  * [New features in version 2](#new-features-in-version-2)
  * [Changes from version 1](#changes-from-version-1)

## Overview

This document describes what's new in chezmoi v2 for chezmoi v1 users.

If you're not using a dotfile manager, then read [why you should you use a
dotfile
manager](https://github.com/twpayne/chezmoi/blob/master/docs/COMPARISON.md#why-should-i-use-a-dotfile-manager)
and then [install
chezmoi](https://github.com/twpayne/chezmoi/blob/master/docs/INSTALL.md) and
follow the [quick-start
guide](https://github.com/twpayne/chezmoi/blob/master/docs/QUICKSTART.md).

If you're already using a dotfile manager, but not chezmoi, then read about [how
chezmoi compares to other dotfile
managers](https://github.com/twpayne/chezmoi/blob/master/docs/COMPARISON.md)
first and then [install
chezmoi](https://github.com/twpayne/chezmoi/blob/master/docs/INSTALL.md), follow
the [quick-start
guide](https://github.com/twpayne/chezmoi/blob/master/docs/QUICKSTART.md), and
read the [how-to
guide](https://github.com/twpayne/chezmoi/blob/master/docs/HOWTO.md) to quickly
discover chezmoi's functionality.

## Version 2

chezmoi version 2 brings many new features and fixes a few corner-case bugs.
Very few, if any, changes should be required to your source directory,
templates, or config file.

### New features in version 2

* The new `chezmoi status` command shows you a concise list of differences, much
  like `git status`.
* The `chezmoi apply` command prompts you if a file has been modified by
  something other than chezmoi since chezmoi last wrote it, keeping you in full
  control of your changes, wherever you make them.
* The `chezmoi init` command will try to guess your dotfile repository if you
  give it a short argument. For example, `chezmoi init username` is the
  equivalent of `chezmoi init https://github.com/username/dotfiles.git`.
* chezmoi includes a builtin `git` command which it will use if it cannot find
  `git` in your `$PATH`. This means that you don't even have to install `git` to
  setup your dotfiles on a new machine.
* chezmoi detects when your config file template has changed and prompts you to
  re-run `chezmoi init` if needed.
* The new `create_` attribute allows you to create a file with initial content,
  but not have it overwritten by `chezmoi apply`.
* The new `modify_` attribute allows you to modify an existing file with a
  script, so you can use chezmoi to manage parts, but not all, of a file.
* The new script attributes `before_` and `after_` control when scripts are run
  relative to when your files are updated.
* The new `--exclude` option allows you to control what types of target will be
  updated. For example `chezmoi apply --exclude=scripts` will cause chezmoi to
  apply everything except scripts and `chezmoi init --apply --exclude=encrypted`
  will exclude encrypted files.
* The new `--keep-going` option causes chezmoi to keep going as far as possible
  rather than stopping at the first error it encounters.
* The new `--no-tty` option stops chezmoi from opening a TTY to read input
  (including passwords) and instead reads them from stdin.
* The new `--source-path` option allows you to specify targets by source path,
  which you can use in an on-save editor hook to automatically run `chezmoi
  apply` when you edit a dotfile in your source state.
* The new `gitHubKeys` template function allows you to populate your
  `~/.ssh/authorized_keys` from your public SSH keys on GitHub.
* The `promptBool` function now also recognizes `y`, `yes`, `on`, `n`, `no`, and
  `off` as boolean values.
* The `chezmoi archive` command now includes scripts in the generated archive,
  and can generate `.zip` files.
* The new `edit.command` and `edit.args` configuration variables give you more
  control over the command invoked by `chezmoi edit`.
* The `chezmoi init` command has a new `--one-shot` option which does a shallow
  clone of your dotfiles repo, runs `chezmoi apply`, and then removes your
  source and configuration directories. It's the fastest way to set up your
  dotfiles on a ephemeral machine and then remove all traces of chezmoi.
* Standard template variables are set on a best-effort basis. If errors are
  encountered, chezmoi leaves the variable unset rather than terminating with
  the error.
* The new `.chezmoi.version` template variable contains the version of chezmoi.
  You can compare versions using [version comparison
  functions](https://masterminds.github.io/sprig/semver.html).
* The new `.chezmoi.fqdnHostname` template variables contains the
  fully-qualified domain name of the machine, if it can be determined.
* You can now encrypt whole files with [age](https://age-encryption.org).

### Changes from version 1

chezmoi version 2 includes a few minor changes from version 1, mainly to enable
the new functionality and for consistency:

* chezmoi uses a different format to persist its state. Specifically, this means
  that all your `run_once_` scripts will be run again the first time you run
  `chezmoi apply`.
* `chezmoi add`, and many other commands, are now recursive by default.
* `chezmoi apply` will warn if a file in the destination directory has been
  modified since chezmoi last wrote it. To force overwriting, pass the `--force`
  option.
* `chezmoi edit` no longer supports the `--prompt` option.
* The only diff format is now `git`. The `diff.format` configuration variable is
  ignored.
* Diffs include the contents of scripts that would be run.
* Mercurial support has been removed.
* The `chezmoi source` command has been removed, use `chezmoi git` instead.
* The `sourceVCS` configuration group has been renamed to `git`.
* The order of files for a three-way merge passed to `merge.command` is now
  actual file, target state, source state.
* The `chezmoi keyring` command has been moved to `chezmoi secret keyring`.
* The `genericSecret` configuration group has been renamed to `secret`.
* The `chezmoi chattr` command uses `encrypted` instead of `encrypt` as the
  attribute for encrypted files.
* The gpg recipient is configured with the `gpg.recipient` configuration
  variable, `gpgRecipient` is no longer used.
* The structure of data output by `chezmoi dump` has changed.
* The `.chezmoi.homedir` template variable has been replaced with
  `.chezmoi.homeDir`. For compatibility, `.chezmoi.homedir` will continue to be
  supported until version 2.1.
* The `.chezmoi.fullHostname` template variable has been removed, as it did not
  contain the full hostname. Use `.chezmoi.fqdnHostname` (UNIX only) instead.
