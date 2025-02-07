# Why use chezmoi?

## Why should I use a dotfile manager?

Dotfile managers give you the combined benefit of a consistent environment
everywhere with an undo command and a restore from backup.

As the core of our development environments become increasingly standardized
(e.g. using git at both home and work), and we further customize them, at the
same time we increasingly work in ephemeral environments like Docker
containers, virtual machines, and GitHub Codespaces.

In the same way that nobody would use an editor without an undo command, or
develop software without a version control system, chezmoi brings the
investment that you have made in mastering your tools to every environment that
you work in.

## I already have a system to manage my dotfiles, why should I use chezmoi?

!!! quote

    I’ve been using Chezmoi for more than a year now, across at least 3
    computers simultaneously, and I really love it. Most of all, I love how
    fast I can configure a new machine when I use it. In just a couple minutes
    of work, I can kick off a process on a brand-new computer that will set up
    my dotfiles and install all my usual software so it feels like a computer
    I’ve been using for years. I also appreciate features like secrets
    management, which allow me to share my dotfiles while keeping my secrets
    safe. Overall, I love the way Chezmoi fits so perfectly into the niche of
    managing dotfiles.

    — [@mike_kasberg][kasberg]

!!! quote

    I had initially been turned off when I first encountered [chezmoi], because
    [chezmoi] seemed overkill for (what appeared to me) a simple task.

    But the problem of managing a relatively small number of dotfiles across a
    relatively small number of machines with small differences between them and
    keeping them up to date proved to be _MUCH_ more complex than I imagined.
    Copy things around by hand, and then later distributing them via source
    control got hairy very quickly.

    I finally realized all those features were absolutely necessary to manage
    things sanely, and once I took some time to learn how to do things with
    chezmoi, I have never looked back.

    — [njt][njt]

!!! quote

    Regular reminder that chezmoi is the best dotfile manager utility I've used
    and you can too

    — @mbbroberg

If you're using any of the following methods:

* A custom shell script.

* An existing dotfile manager like [dotbot][dotbot], [rcm][rcm],
  [homesick][homesick], [vcsh][vcsh], [yadm][yadm], or [GNU Stow][stow].

* A [bare git repo][baregit].

Then you've probably run into at least one of the following problems.

### ...if coping with differences between machines requires extra effort

If you want to synchronize your dotfiles across multiple operating systems or
distributions, then you may need to manually perform extra steps to cope with
differences from machine to machine. You might need to run different commands on
different machines, maintain separate per-machine files or branches (with the
associated hassle of merging, rebasing, or copying each change), or hope that
your custom logic handles the differences correctly.

chezmoi uses a single source of truth (a single branch) and a single command
that works on every machine. Individual files can be templates to handle machine
to machine differences, if needed.

### ...if you have to keep your dotfiles repo private

!!! quote

    And regarding dotfiles, I saw that. It's only public dotfiles repos so I
    have to evaluate my dotfiles history to be sure. I have secrets scanning
    and more, but it was easier to keep it private for security, I'm ok mostly
    though. I'm using chezmoi and it's easier now

    — @sheldon_hull

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

### ...if you have to maintain your own tool

!!! quote

    I've offloaded my dotfiles deployment from a homespun shell script to chezmoi.
    I'm quite happy with this decision.

    — @gotgenes

!!! quote

    I discovered chezmoi and it's pretty cool, just migrated my old custom
    multi-machine sync dotfile setup and it's so much simpler now

    in case you're wondering I have written 0 code

    — @buritica

!!! quote

    Chezmoi is like what you might get if you re-wrote my bash script in Go,
    came up with better solutions than `diff` for managing config on multiple
    machines, added in secrets management and other useful dotfile tools, and
    tweaked and perfected it over years.

    - [@mike_kasberg][kasberg]

If your system was written by you for your personal use, then it probably has
the functionality that you needed when you wrote it. If you need more
functionality then you have to implement it yourself.

chezmoi includes a huge range of battle-tested functionality out-of-the-box,
including dry-run and diff modes, script execution, conflict resolution, Windows
support, and much, much more. chezmoi is [used by thousands of people][stars]
and has a rich suite of both unit and integration tests. When you hit the limits
of your existing dotfile management system, chezmoi already has
a tried-and-tested solution ready for you to use.

### ...if setting up your dotfiles requires more than one short command

If your system is written in a scripting language like Python, Perl, or Ruby,
then you also need to install a compatible version of that language's runtime
before you can use your system.

chezmoi is distributed as a single stand-alone statically-linked binary with no
dependencies that you can simply copy onto your machine and run. You don't even
need git installed. chezmoi provides one-line installs, pre-built binaries,
packages for Linux and BSD distributions, Homebrew formulae, Scoop and
Chocolatey support on Windows, and a initial config file generation mechanism to
make installing your dotfiles on a new machine as painless as possible.

[baregit]: https://www.atlassian.com/git/tutorials/dotfiles
[dotbot]: https://github.com/anishathalye/dotbot
[homesick]: https://github.com/technicalpickles/homesick
[kasberg]: https://www.mikekasberg.com/blog/2021/05/12/my-dotfiles-story.html
[njt]: https://news.ycombinator.com/item?id=31015669
[rcm]: https://github.com/thoughtbot/rcm
[stars]: https://github.com/twpayne/chezmoi/stargazers
[stow]: https://www.gnu.org/software/stow/
[vcsh]: https://github.com/RichiH/vcsh
[yadm]: https://yadm.io/
