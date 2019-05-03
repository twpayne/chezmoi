# chezmoi

[![Build
Status](https://travis-ci.org/twpayne/chezmoi.svg?branch=master)](https://travis-ci.org/twpayne/chezmoi)
[![Report
Card](https://goreportcard.com/badge/github.com/twpayne/chezmoi)](https://goreportcard.com/report/github.com/twpayne/chezmoi)
[![Coverage Status](https://coveralls.io/repos/github/twpayne/chezmoi/badge.svg)](https://coveralls.io/github/twpayne/chezmoi)

Manage your dotfiles across multiple machines, securely.

* [What does chezmoi do and why should I use it?](#what-does-chezmoi-do-and-why-should-i-use-it)
* [Features](#features)
* [I already have a system to manage my dotfiles, why should I use chezmoi?](#i-already-have-a-system-to-manage-my-dotfiles-why-should-i-use-chezmoi)
* [Documentation](#documentation)
* [Related projects](#related-projects)
* [License](#license)

## What does chezmoi do and why should I use it?

chezmoi helps you manage your personal configuration files (dotfiles) across
multiple machines. It is particularly helpful if you have spent time customizing
the tools you use (e.g. shells, editors, and version control systems) and want
to keep machines running different accounts (e.g. home and work) and/or
different operating systems (e.g. Linux and macOS) in sync, while still be able
to easily cope with differences from machine to machine.

chezmoi has particularly strong support for security features, allowing you to
manage secrets (e.g. passwords, access tokens, and private keys) securely and
seamlessly using either gpg encryption or a password manager of your choice.

If you do not personalize your configuration or only ever use a single operating
system with a single account then you don't need chezmoi. Otherwise, read on...

## Features

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
  [Bitwarden](https://bitwarden.com/), [LastPass](https://lastpass.com/),
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
  day-to-day operations one line commands.

## I already have a system to manage my dotfiles, why should I use chezmoi?

* If your system is based on copying files with a shell script or creating
  symlinks (e.g. using [GNU
  Stow](http://brandon.invergo.net/news/2012-05-26-using-gnu-stow-to-manage-your-dotfiles.html))
  then handling files that vary from machine to machine requires manual work.
  You might need to maintain separate config files for separate machines, or run
  different commands on different machines. chezmoi gives you a single command
  that works on every machine.

* If your system is based on using git with a different branches for different
  machines, then you need manually merge or rebase to ensure that changes you
  make are applied to each machine. chezmoi uses a single branch and makes it
  trivial to share common parts while allowing specific per-machine
  configuration.

* If your system stores secrets in plain text, then your dotfiles repository
  must be private. With chezmoi you never need to store secrets in your
  repository, so you can make it public. You can check out your repository on
  your work machine and not fear that this will give your work IT department
  access to your personal data.

* If your system was written by you for your personal use, then it probably has
  the minimum functionality that you need. chezmoi includes a wide range of
  functionality out-of-the-box, including dry run and diff modes.

* All systems suffer from the "bootstrap" problem: you need to install your
  system before you can install your dotfiles. chezmoi provides one-line
  installs, statically-linked binaries, packages for many Linux and BSD
  distributions, Homebrew formulae, and a initial config file generation
  mechanism to make overcoming the bootstrap problem as painless as possible.

## Documentation

chezmoi includes five types of documentation:

* An [installation guide](docs/INSTALL.md).
* A [quick start guide](docs/QUICKSTART.md).
* A [how-to guide](docs/HOWTO.md) for achieving specific tasks.
* An [FAQ](docs/FAQ.md) for questions that aren't answered elsewhere.
* A [reference](docs/REFERENCE.md) for a complete description of chezmoi.

## Related projects

See [`dotfiles.github.io`](https://dotfiles.github.io/).

## License

MIT
