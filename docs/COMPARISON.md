# chezmoi Comparison guide

<!--- toc --->
* [Go to chezmoi.io](#go-to-chezmoiio)
* [Comparison table](#comparison-table)
* [I already have a system to manage my dotfiles, why should I use chezmoi?](#i-already-have-a-system-to-manage-my-dotfiles-why-should-i-use-chezmoi)
  * [...if coping with differences between machines requires special care](#if-coping-with-differences-between-machines-requires-special-care)
  * [...if you need to think for a moment before giving anyone access to your dotfiles](#if-you-need-to-think-for-a-moment-before-giving-anyone-access-to-your-dotfiles)
  * [...if your needs are outgrowing your current tool](#if-your-needs-are-outgrowing-your-current-tool)
  * [...if setting up your dotfiles requires more than two short commands](#if-setting-up-your-dotfiles-requires-more-than-two-short-commands)

## Go to chezmoi.io

You are looking at documentation for chezmoi version 2, which hasn't been
released yet. Documentation for the current version of chezmoi is at
[chezmoi.io](https://chezmoi.io/docs/comparison/).
## Comparison table

[chezmoi]: https://chezmoi.io/
[dotbot]: https://github.com/anishathalye/dotbot
[rcm]: https://github.com/thoughtbot/rcm
[homesick]: https://github.com/technicalpickles/homesick
[yadm]: https://yadm.io/
[bare git]: https://www.atlassian.com/git/tutorials/dotfiles "bare git"

|                                        | [chezmoi]     | [dotbot]          | [rcm]             | [homesick]        | [yadm]        | [bare git] |
| -------------------------------------- | ------------- | ----------------- | ----------------- | ----------------- | ------------- | ---------- |
| Implementation language                | Go            | Python            | Perl              | Ruby              | Bash          | C          |
| Distribution                           | Single binary | Python package    | Multiple files    | Ruby gem          | Single script | n/a        |
| Install method                         | Multiple      | git submodule     | Multiple          | Ruby gem          | Multiple      | n/a        |
| Non-root install on bare system        | Yes           | Difficult         | Difficult         | Difficult         | Yes           | Yes        |
| Windows support                        | Yes           | No                | No                | No                | No            | Yes        |
| Bootstrap requirements                 | None          | Python, git       | Perl, git         | Ruby, git         | git           | git        |
| Source repos                           | Single        | Single            | Multiple          | Single            | Single        | Single     |
| Method                                 | File          | Symlink           | File              | Symlink           | File          | File       |
| Config file                            | Optional      | Required          | Optional          | None              | None          | No         |
| Private files                          | Yes           | No                | No                | No                | No            | No         |
| Show differences without applying      | Yes           | No                | No                | No                | Yes           | Yes        |
| Whole file encryption                  | Yes           | No                | No                | No                | Yes           | No         |
| Password manager integration           | Yes           | No                | No                | No                | No            | No         |
| Machine-to-machine file differences    | Templates     | Alternative files | Alternative files | Alternative files | Templates     | Manual     |
| Custom variables in templates          | Yes           | n/a               | n/a               | n/a               | No            | No         |
| Executable files                       | Yes           | Yes               | Yes               | Yes               | No            | Yes        |
| File creation with initial contents    | Yes           | No                | No                | No                | No            | No         |
| Manage partial files                   | Yes           | No                | No                | No                | No            | No         |
| File removal                           | Yes           | Manual            | No                | No                | No            | No         |
| Directory creation                     | Yes           | Yes               | Yes               | No                | No            | Yes        |
| Run scripts                            | Yes           | Yes               | Yes               | No                | No            | No         |
| Run once scripts                       | Yes           | No                | No                | No                | Manual        | No         |
| Machine-to-machine symlink differences | Yes           | No                | No                | No                | Yes           | No         |
| Shell completion                       | Yes           | No                | No                | No                | Yes           | Yes        |
| Archive import                         | Yes           | No                | No                | No                | No            | No         |
| Archive export                         | Yes           | No                | No                | No                | No            | Yes        |

## I already have a system to manage my dotfiles, why should I use chezmoi?

If you're using any of the following methods:

* A custom shell script.
* An existing dotfile manager like
  [homeshick](https://github.com/andsens/homeshick),
  [homesick](https://github.com/technicalpickles/homesick),
  [rcm](https://github.com/thoughtbot/rcm), [GNU
  Stow](https://www.gnu.org/software/stow/), or [yadm](https://yadm.io/).
* A [bare git repo](https://www.atlassian.com/git/tutorials/dotfiles).

Then you've probably run into at least one of the following problems.

### ...if coping with differences between machines requires special care

If you want to synchronize your dotfiles across multiple operating systems or
distributions, then you may need to manually perform extra steps to cope with
differences from machine to machine. You might need to run different commands on
different machines, maintain separate per-machine files or branches (with the
associated hassle of merging, rebasing, or copying each change), or hope that
your custom logic handles the differences correctly.

chezmoi uses a single source of truth (a single branch) and a single command
that works on every machine. Individual files can be templates to handle machine
to machine differences, if needed.

### ...if you need to think for a moment before giving anyone access to your dotfiles

If your system stores secrets in plain text, then you must be very careful about
where you clone your dotfiles. If you clone them on your work machine then
anyone with access to your work machine (e.g. your IT department) will have
access to your home secrets. If you clone it on your home machine then you risk
leaking work secrets.

With chezmoi you can store secrets in your password manager or encrypt them, and
even store passwords in different ways on different machines. You can clone your
dotfiles repository anywhere, and even make your dotfiles repo public, without
leaving personal secrets on your work machine or work secrets on your personal
machine.

### ...if your needs are outgrowing your current tool

If your system was written by you for your personal use, then it probably has
the minimum functionality that you needed when you wrote it. If you need more
functionality then you have to implement it yourself.

chezmoi includes a huge range of battle-tested functionality out-of-the-box,
including dry-run and diff modes, script execution, conflict resolution, Windows
support, and much, much more. chezmoi is [used by thousands of
people](https://github.com/twpayne/chezmoi/stargazers), so it is likely that
when you hit the limits of your existing dotfile management system, chezmoi
already has a tried-and-tested solution ready for you to use.

### ...if setting up your dotfiles requires more than two short commands

If your system is written in a scripting language like Python, Perl, or Ruby,
then you also need to install a compatible version of that language's runtime
before you can use your system.

chezmoi is distributed as a single stand-alone statically-linked binary with no
dependencies that you can simply copy onto your machine and run. chezmoi
provides one-line installs, pre-built binaries, packages for Linux and BSD
distributions, Homebrew formulae, Scoop and Chocolatey support on Windows, and a
initial config file generation mechanism to make installing your dotfiles on a
new machine as painless as possible.

