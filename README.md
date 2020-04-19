# ![chezmoi logo](logo-144px.svg) chezmoi

[![GitHub Release](https://img.shields.io/github/release/twpayne/chezmoi.svg)](https://github.com/twpayne/chezmoi/releases)

Manage your dotfiles across multiple machines, securely.

* [Documentation](#documentation)
* [What does chezmoi do and why should I use it?](#what-does-chezmoi-do-and-why-should-i-use-it)
* [What are chezmoi's key features?](#what-are-chezmois-key-features)
* [I already have a system to manage my dotfiles, why should I use chezmoi?](#i-already-have-a-system-to-manage-my-dotfiles-why-should-i-use-chezmoi)
* [What does a chezmoi dotfile repo look like?](#what-does-a-chezmoi-dotfile-repo-look-like)
* [How do I start with chezmoi?](#how-do-i-start-with-chezmoi)
* [License](#license)

## Documentation

* [Install guide](docs/INSTALL.md) to get chezmoi installed on your machine with
  one or two commands.
* [Quick start guide](docs/QUICKSTART.md) for your first steps.
* [How-to guide](docs/HOWTO.md) for achieving specific tasks.
* [FAQ](docs/FAQ.md) for questions that aren't answered elsewhere.
* [Changes](docs/CHANGES.md) for non-backwards compatible changes.
* [Reference](docs/REFERENCE.md) for a complete description of chezmoi.
* [Contributing](docs/CONTRIBUTING.md) for people looking to contribute to or
  package chezmoi.

## What does chezmoi do and why should I use it?

chezmoi helps you manage your personal configuration files (dotfiles, like
`~/.bashrc`) across multiple machines.

chezmoi is helpful if you have spent time customizing the tools you use (e.g.
shells, editors, and version control systems) and want to keep machines running
different accounts (e.g. home and work) and/or different operating systems (e.g.
Linux and macOS) in sync, while still being able to easily cope with differences
from machine to machine.

chezmoi has strong support for security, allowing you to manage secrets (e.g.
passwords, access tokens, and private keys) securely and seamlessly using a
password manager of your choice or gpg encryption.

In all cases you only need to maintain a single source of truth: a single branch
in a version control system (e.g. git) for everything public and a single
password manager for all your secrets.

If you do not personalize your configuration or only ever use a single operating
system with a single account and none of your dotfiles contain secrets then you
don't need chezmoi. Otherwise, read on...

## What are chezmoi's key features?

* Flexible: You can share as much configuration across machines as you want,
  while still being able to control machine-specific details. You only need to
  maintain a single branch. Your dotfiles can be templates (using
  [`text/template`](https://pkg.go.dev/text/template) syntax). Predefined
  variables allow you to change behaviour depending on operating system,
  architecture, and hostname.

* Personal and secure: Nothing leaves your machine, unless you want it to. You
  can use the version control system of your choice to manage your
  configuration, and you can write the configuration file in the format of your
  choice. chezmoi can retrieve secrets from [1Password](https://1password.com/),
  [Bitwarden](https://bitwarden.com/), [gopass](https://www.gopass.pw/),
  [KeePassXC](https://keepassxc.org/), [LastPass](https://lastpass.com/),
  [pass](https://www.passwordstore.org/), [Vault](https://www.vaultproject.io/),
  Keychain, [Keyring](https://wiki.gnome.org/Projects/GnomeKeyring), or any
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
  links in your source of truth and chezmoi updates your home directory to match
  that state. What you want is what you get.

* Fast and easy to use: chezmoi runs in fractions of a second and makes most
  day-to-day operations one line commands, including installation,
  initialization, and keeping your machines up-to-date. chezmoi can pull and
  apply changes from your dotfiles repo in a single command, and automatically
  commit and push changes.

## I already have a system to manage my dotfiles, why should I use chezmoi?

If you're using any of the following methods:

* A custom shell script.
* An existing dotfile manager like
  [homeshick](https://github.com/andsens/homeshick),
  [homesick](https://github.com/technicalpickles/homesick),
  [rcm](https://github.com/thoughtbot/rcm), [GNU
  Stow](https://www.gnu.org/software/stow/), or [yadm](https://yadm.io/).
* A [bare git repo](https://www.atlassian.com/git/tutorials/dotfiles).

Then you've probably run into at least one of the following problems:

* If you want to synchronize your dotfiles across multiple operating systems or
  distributions, then you may need to manually perform extra steps to cope with
  differences from machine to machine. You might need to run different commands
  on different machines, maintain separate per-machine branches or files (with
  the associated hassle of merging, rebasing, or copying each change), or hope
  that your custom logic handles the differences correctly. chezmoi uses a
  single source of truth (a single branch) and a single command that works on
  every machine. Individual files can be templates to handle machine to machine
  differences, if needed.

* If your system stores secrets in plain text, then you must be very careful
  about where you clone it. If you clone it on your work machine then anyone
  with access to your work machine (e.g. your IT department) will have access to
  your home secrets. If you clone it on your home machine then you risk leaking
  work secrets. With chezmoi you can store secrets in your password manager or
  encrypt them, and even use different password managers or encryption keys for
  your work and home machines. You can clone your repository on every machine,
  and even make your dotfiles repo public, without leaving personal secrets on
  your work machine or work secrets on your personal machine.

* If your system was written by you for your personal use, then it probably has
  the minimum functionality that you need. You might need special logic to
  handle dotfiles that need to be private or run configuration scripts
  occasionally. chezmoi includes a huge range of battle-tested functionality
  out-of-the-box, including dry-run and diff modes, script execution, conflict
  resolution, Windows support, and much, much more. chezmoi is [used by
  thousands of people](https://github.com/twpayne/chezmoi/stargazers), so it is
  likely that when you hit the limits of your existing dotfile management
  system, chezmoi already has a tried-and-tested solution ready to use.

* If your system is written in a scripting language like Python, Perl, or Ruby,
  then you also need to install that language's runtime before you can use your
  system. chezmoi is distributed as a single stand-alone statically-linked
  binary with no dependencies that you can simply copy onto your machine and
  run. chezmoi provides one-line installs, pre-built binaries, packages for
  Linux and BSD distributions, Homebrew formulae, Scoop support on Windows, and
  a initial config file generation mechanism to make installing your dotfiles on
  a new machine as painless as possible.

## What does a chezmoi dotfile repo look like?

Have a look at [repos tagged with `chezmoi` on
GitHub](https://github.com/topics/chezmoi?o=desc&s=updated). You can also read
what [has been written about chezmoi](docs/MEDIA.md).

## How do I start with chezmoi?

[Install chezmoi](docs/INSTALL.md) then read the [quick start
guide](docs/QUICKSTART.md). The [how-to guide](docs/HOWTO.md) covers most common
tasks, and there's the [frequently asked questions](docs/FAQ.md) for specific
questions. You can browse other people's [dotfiles that use
chezmoi](https://github.com/topics/chezmoi). For a full description of chezmoi,
consult the [reference](docs/REFERENCE.md). If all else fails, [open an
issue](https://github.com/twpayne/chezmoi/issues/new/choose).

## License

MIT
