# What does chezmoi do?

chezmoi helps you manage your personal configuration files (dotfiles, like
`~/.gitconfig`) across multiple machines.

chezmoi is helpful if you have spent time customizing the tools you use (e.g.
shells, editors, and version control systems) and want to keep machines running
different accounts (e.g. home and work) and/or different operating systems
(e.g. Linux, macOS, and Windows) in sync, while still being able to easily cope
with differences from machine to machine.

chezmoi scales from the trivial (e.g. copying a few dotfiles onto a Raspberry
Pi, development container, or virtual machine) to complex long-lived
multi-machine development environments (e.g. keeping any number of home and
work, Linux, macOS, and Windows machines in sync). In all cases you only need
to maintain a single source of truth (a single branch in git) and getting
started only requires adding a single binary to your machine (which you can do
with `curl`, `wget`, or `scp`).

chezmoi has strong support for security, allowing you to manage secrets (e.g.
passwords, access tokens, and private keys) securely and seamlessly using a
password manager and/or encrypt whole files with your favorite encryption tool.

If you do not personalize your configuration or only ever use a single
operating system with a single account and none of your dotfiles contain
secrets then you don't need chezmoi. Otherwise, read on...

## What are chezmoi's key features?

### Flexible

You can share as much configuration across machines as you want, while still
being able to control machine-specific details.Your dotfiles can be templates
(using [`text/template`](https://pkg.go.dev/text/template) syntax). Predefined
variables allow you to change behavior depending on operating system,
architecture, and hostname. chezmoi runs on all commonly-used platforms, like
Linux, macOS, and Windows. It also runs on less commonly-used platforms, like
FreeBSD, OpenBSD, and Termux.

### Personal and secure

Nothing leaves your machine, unless you want it to. Your configuration remains
in a git repo under your control. You can write the configuration file in the
format of your choice. chezmoi can retrieve secrets from
[1Password](https://1password.com/), [AWS Secrets
Manager](https://aws.amazon.com/secrets-manager/),
[Bitwarden](https://bitwarden.com/), [Dashlane](https://www.dashlane.com/),
[gopass](https://www.gopass.pw/), [HCP Vault
Secrets](https://developer.hashicorp.com/hcp/docs/vault-secrets),
[KeePassXC](https://keepassxc.org/), [Keeper](https://www.keepersecurity.com/),
[LastPass](https://lastpass.com/), [pass](https://www.passwordstore.org/),
[passhole](https://github.com/Evidlo/passhole),
[Vault](https://www.vaultproject.io/), Keychain,
[Keyring](https://wiki.gnome.org/Projects/GnomeKeyring), or any command-line
utility of your choice. You can encrypt individual files with
[GnuPG](https://www.gnupg.org) or [age](https://age-encryption.org). You can
checkout your dotfiles repo on as many machines as you want without revealing
any secrets to anyone.

### Transparent

chezmoi includes verbose and dry run modes so you can review exactly what
changes it will make to your home directory before making them. chezmoi's
source format uses only regular files and directories that map one-to-one with
the files, directories, and symlinks in your home directory that you choose to
manage. If you decide not to use chezmoi in the future, it is easy to move your
data elsewhere.

### Declarative and robust

You declare the desired state of files, directories, and symbolic links in your
source of truth and chezmoi updates your home directory to match that state.
What you want is what you get. chezmoi updates all files and symbolic links
atomically. You will never be left with incomplete files that could lock you
out, even if the update process is interrupted.

### Fast and easy to use

Using chezmoi feels like using git: the commands are similar and chezmoi runs
in fractions of a second. chezmoi makes most day-to-day operations one line
commands, including installation, initialization, and keeping your machines
up-to-date. chezmoi can pull and apply changes from your dotfiles repo in a
single command, and automatically commit and push changes.
