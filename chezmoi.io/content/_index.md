---
title: Home
type: docs
---

# chezmoi

Manage your dotfiles across multiple machines, securely.

## What does chezmoi do and why should I use it?

chezmoi helps you manage your personal configuration files (dotfiles) across
multiple machines. It is particularly helpful if you have spent time customizing
the tools you use (e.g. shells, editors, and version control systems) and want
to keep machines running different accounts (e.g. home and work) and/or
different operating systems (e.g. Linux and macOS) in sync, while still being
able to easily cope with differences from machine to machine.

chezmoi has particularly strong support for security, allowing you to manage
secrets (e.g. passwords, access tokens, and private keys) securely and
seamlessly using either gpg encryption or a password manager of your choice.

In all cases you only need to maintain a single source of truth: a single branch
in a version control system (e.g. git) for everything public and a single
password manager for all your secrets, with seamless integration between them.

If you do not personalize your configuration or only ever use a single operating
system with a single account and none of your dotfiles contain secrets then you
don't need chezmoi. Otherwise, read on...

## What are chezmoi's key features?

* Flexible: You can share as much configuration across machines as you want,
  while still being able to control machine-specific details. You only need to
  maintain a single branch. Your dotfiles can be templates (using
  [`text/template`](https://godoc.org/text/template) syntax). Predefined
  variables allow you to change behaviour depending on operating system,
  architecture, and hostname.

* Personal and secure: Nothing leaves your machine, unless you want it to. You
  can use the version control system of your choice to manage your
  configuration, and you can write the configuration file in the format of your
  choice. chezmoi can retrieve secrets from [1Password](https://1password.com/),
  [Bitwarden](https://bitwarden.com/), [gopass](https://www.gopass.pw/),
  [KeePassXC](https://keepassxc.org/), [LastPass](https://lastpass.com/),
  [pass](https://www.passwordstore.org/), [Vault](https://www.vaultproject.io/),
  your Keychain (on macOS), [GNOME
  Keyring](https://wiki.gnome.org/Projects/GnomeKeyring) (on Linux), or any
  command-line utility of your choice. You can encrypt individual files with
  [gpg](https://www.gnupg.org). You can checkout your dotfiles repo on as many
  machines as you want without revealing any secrets to anyone.

* Transparent: chezmoi includes verbose and dry run modes so you can review
  exactly what changes it will make to your home directory before making them.
  chezmoi's source format uses only regular files and directories that map
  one-to-one with the files, directories, and symlinks in your home directory
  that you choose to manage. If you decide not to use chezmoi in the future, it
  is easy to move your data elsewhere.

* Robust: chezmoi updates all files and symbolic links atomically (using
  [`google/renameio`](https://github.com/google/renameio)). You will never be
  left with incomplete files that could lock you out, even if the update process
  is interrupted.

* Declarative: you declare the desired state of files, directories, and symbolic
  links in your home directory and chezmoi updates your home directory to match
  that state. What you want is what you get.

* Fast and easy to use: chezmoi runs in fractions of a second and makes most
  day-to-day operations one line commands, including installation,
  initialization, and keeping your machines up-to-date.

## I already have a system to manage my dotfiles, why should I use chezmoi?

* If your system is based on copying files with a shell script or creating
  symlinks (e.g. using [GNU
  Stow](http://brandon.invergo.net/news/2012-05-26-using-gnu-stow-to-manage-your-dotfiles.html))
  then handling files that vary from machine to machine requires manual work.
  You might need to maintain separate config files for separate machines, or run
  different commands on different machines. chezmoi gives you a single command
  (`chezmoi update`)  that works on every machine.

* If your system is based on using git with a different branches for different
  machines, then you need manually merge or rebase to ensure that changes you
  make are applied to each machine. chezmoi uses a single branch and makes it
  trivial to share common parts while allowing specific per-machine
  configuration.

* If your system stores secrets in plain text, then your dotfiles repository
  must be private. With chezmoi you can store secrets in your password manager,
  so you can make your dotfiles public. You can share your repository between
  your personal and work machines, without leaving personal secrets on your work
  machine or work secrets on your personal machine.

* If your system was written by you for your personal use, then it probably has
  the minimum functionality that you need. chezmoi includes a wide range of
  functionality out-of-the-box, including dry run and diff modes, conflict
  resolution, and running scripts.

* All systems suffer from the bootstrap problem: you need to install your system
  before you can install your dotfiles. chezmoi provides one-line installs,
  statically-linked binaries, packages for many Linux and BSD distributions,
  Homebrew formulae, and a initial config file generation mechanism to make
  overcoming the bootstrap problem as painless as possible.

## How do I start with chezmoi?

[Install chezmoi](/docs/install/) then read the [quick start
guide](/docs/quick-start/). The [how-to guide](/docs/how-to/) covers most common
tasks, and there's the [frequently asked questions](docs/faq/) for specific
questions. For a full description of chezmoi, consult the
[reference](/docs/reference/). If all else fails, [open an
issue](https://github.com/twpayne/chezmoi/issues/new).

## License

MIT
