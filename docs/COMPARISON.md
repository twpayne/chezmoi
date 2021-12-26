# chezmoi comparison guide

<!--- toc --->
* [Comparison table](#comparison-table)
* [Why should I use a dotfile manager?](#why-should-i-use-a-dotfile-manager)
* [I already have a system to manage my dotfiles, why should I use chezmoi?](#i-already-have-a-system-to-manage-my-dotfiles-why-should-i-use-chezmoi)
  * [Coping with differences between machines requires extra effort](#coping-with-differences-between-machines-requires-extra-effort)
  * [You have to keep your dotfiles repo private](#you-have-to-keep-your-dotfiles-repo-private)
  * [You have to maintain your own tool](#you-have-to-maintain-your-own-tool)
  * [Setting up your dotfiles requires more than one short command](#setting-up-your-dotfiles-requires-more-than-one-short-command)

---

## Comparison table

[chezmoi]: https://chezmoi.io/
[dotbot]: https://github.com/anishathalye/dotbot
[rcm]: https://github.com/thoughtbot/rcm
[homesick]: https://github.com/technicalpickles/homesick
[vcsh]: https://github.com/RichiH/vcsh
[yadm]: https://yadm.io/
[bare git]: https://www.atlassian.com/git/tutorials/dotfiles "bare git"

|                                        | [chezmoi]     | [dotbot]          | [rcm]             | [homesick]        | [vcsh]                   | [yadm]        | [bare git] |
| -------------------------------------- | ------------- | ----------------- | ----------------- | ----------------- | ------------------------ | ------------- | ---------- |
| Distribution                           | Single binary | Python package    | Multiple files    | Ruby gem          | Single script or package | Single script | -          |
| Install method                         | Many          | git submodule     | Many              | Ruby gem          | Many                     | Many          | Manual     |
| Non-root install on bare system        | ✅            | ⁉️                 | ⁉️                 | ⁉️                 | ✅                       | ✅            | ✅         |
| Windows support                        | ✅            | ❌                | ❌                | ❌                | ❌                       | ✅            | ✅         |
| Bootstrap requirements                 | None          | Python, git       | Perl, git         | Ruby, git         | sh, git                  | git           | git        |
| Source repos                           | Single        | Single            | Multiple          | Single            | Multiple                 | Single        | Single     |
| dotfiles are...                        | Files         | Symlinks          | Files             | Symlinks          | Files                    | Files         | Files      |
| Config file                            | Optional      | Required          | Optional          | None              | None                     | Optional      | Optional   |
| Private files                          | ✅            | ❌                | ❌                | ❌                | ❌                       | ✅            | ❌         |
| Show differences without applying      | ✅            | ❌                | ❌                | ❌                | ✅                       | ✅            | ✅         |
| Whole file encryption                  | ✅            | ❌                | ❌                | ❌                | ❌                       | ✅            | ❌         |
| Password manager integration           | ✅            | ❌                | ❌                | ❌                | ❌                       | ❌            | ❌         |
| Machine-to-machine file differences    | Templates     | Alternative files | Alternative files | Alternative files | Branches                 | Alternative files, Templates     | ⁉️          |
| Custom variables in templates          | ✅            | ❌                | ❌                | ❌                | ❌                       | ❌            | ❌         |
| Executable files                       | ✅            | ✅                | ✅                | ✅                | ✅                       | ✅            | ✅         |
| File creation with initial contents    | ✅            | ❌                | ❌                | ❌                | ✅                       | ❌            | ❌         |
| Externals                              | ✅            | ❌                | ❌                | ❌                | ❌                       | ❌            | ❌         |
| Manage partial files                   | ✅            | ❌                | ❌                | ❌                | ⁉️                        | ✅            | ⁉️          |
| File removal                           | ✅            | ❌                | ❌                | ❌                | ✅                       | ✅            | ❌         |
| Directory creation                     | ✅            | ✅                | ✅                | ❌                | ✅                       | ❌            | ✅         |
| Run scripts                            | ✅            | ✅                | ✅                | ❌                | ✅                       | ❌            | ❌         |
| Run once scripts                       | ✅            | ❌                | ❌                | ❌                | ✅                       | ❌            | ❌         |
| Machine-to-machine symlink differences | ✅            | ❌                | ❌                | ❌                | ⁉️                        | ✅            | ⁉️          |
| Shell completion                       | ✅            | ❌                | ❌                | ❌                | ✅                       | ✅            | ✅         |
| Archive import                         | ✅            | ❌                | ❌                | ❌                | ✅                       | ❌            | ✅         |
| Archive export                         | ✅            | ❌                | ❌                | ❌                | ✅                       | ❌            | ✅         |
| Implementation language                | Go            | Python            | Perl              | Ruby              | POSIX Shell              | Bash          | C          |

✅ Supported, ⁉️  Possible with significant manual effort, ❌ Not supported

For more comparisons, visit [dotfiles.github.io](https://dotfiles.github.io/).

---

## Why should I use a dotfile manager?

Dotfile managers give you the combined benefit of a consistent environment
everywhere with an undo command and a restore from backup.

As the core of our development environments become increasingly standardized
(e.g. git or Mercurial interfaces to version control at both home and work), and
we further customize them (with shell configs like
[powerlevel10k](https://github.com/romkatv/powerlevel10k)), at the same time we
increasingly work in ephemeral environments like Docker containers and [GitHub
Codespaces](https://github.com/features/codespaces).

chezmoi helps you bring your personal configuration to every environment that
you're working in. In the same way that nobody would use an editor without an
undo command, or develop software without a version control system, chezmoi
brings the investment that you have made in mastering your tools to every
environment that you work in.

---

## I already have a system to manage my dotfiles, why should I use chezmoi?

> Regular reminder that chezmoi is the best dotfile manager utility I've used
> and you can too
>
> — [@mbbroberg](https://twitter.com/mbbroberg/status/1355644967625125892)

If you're using any of the following methods:

* A custom shell script.
* An existing dotfile manager like
  [dotbot](https://github.com/anishathalye/dotbot),
  [rcm](https://github.com/thoughtbot/rcm),
  [homesick](https://github.com/technicalpickles/homesick),
  [vcsh](https://github.com/RichiH/vcsh),
  [yadm](https://yadm.io/), or [GNU Stow](https://www.gnu.org/software/stow/).
* A [bare git repo](https://www.atlassian.com/git/tutorials/dotfiles).

Then you've probably run into at least one of the following problems.

---

### Coping with differences between machines requires extra effort

If you want to synchronize your dotfiles across multiple operating systems or
distributions, then you may need to manually perform extra steps to cope with
differences from machine to machine. You might need to run different commands on
different machines, maintain separate per-machine files or branches (with the
associated hassle of merging, rebasing, or copying each change), or hope that
your custom logic handles the differences correctly.

chezmoi uses a single source of truth (a single branch) and a single command
that works on every machine. Individual files can be templates to handle machine
to machine differences, if needed.

---

### You have to keep your dotfiles repo private

> And regarding dotfiles, I saw that. It's only public dotfiles repos so I have
> to evaluate my dotfiles history to be sure. I have secrets scanning and more,
> but it was easier to keep it private for security, I'm ok mostly though. I'm
> using chezmoi and it's easier now
>
> — [@sheldon_hull](https://twitter.com/sheldon_hull/status/1308139570597371907)

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

---

### You have to maintain your own tool

> I've offloaded my dotfiles deployment from a homespun shell script to chezmoi.
> I'm quite happy with this decision.
>
> — [@gotgenes](https://twitter.com/gotgenes/status/1251008845163319297)

> I discovered chezmoi and it's pretty cool, just migrated my old custom
> multi-machine sync dotfile setup and it's so much simpler now
>
> in case you're wondering I have written 0 code
>
> — [@buritica](https://twitter.com/buritica/status/1361062902451630089)

If your system was written by you for your personal use, then it probably has
the functionality that you needed when you wrote it. If you need more
functionality then you have to implement it yourself.

chezmoi includes a huge range of battle-tested functionality out-of-the-box,
including dry-run and diff modes, script execution, conflict resolution, Windows
support, and much, much more. chezmoi is [used by thousands of
people](https://github.com/twpayne/chezmoi/stargazers) and has a rich suite of
both unit and integration tests. When you hit the limits of your existing
dotfile management system, chezmoi already has a tried-and-tested solution ready
for you to use.

---

### Setting up your dotfiles requires more than one short command

If your system is written in a scripting language like Python, Perl, or Ruby,
then you also need to install a compatible version of that language's runtime
before you can use your system.

chezmoi is distributed as a single stand-alone statically-linked binary with no
dependencies that you can simply copy onto your machine and run. You don't even
need git installed. chezmoi provides one-line installs, pre-built binaries,
packages for Linux and BSD distributions, Homebrew formulae, Scoop and
Chocolatey support on Windows, and a initial config file generation mechanism to
make installing your dotfiles on a new machine as painless as possible.

---
