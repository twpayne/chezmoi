# Customize your source directory

## Use a subdirectory of your dotfiles repo as the root of the source state

By default, chezmoi uses the root of your dotfiles repo as the root of the
source state. If your source state contains many entries in its root, then your
target directory (usually your home directory) will in turn be filled with many
entries in its root as well. You can reduce the number of entries by keeping
`.chezmoiignore` up to date, but this can become tiresome.

Instead, you can specify that chezmoi should read the source state from a
subdirectory of the source directory instead by creating a file called
`.chezmoiroot` containing the relative path to this subdirectory.

For example, given:

``` title="~/.local/share/chezmoi/.chezmoiroot"
home
```

Then chezmoi will read the source state from the `home` subdirectory of your
source directory, for example the desired state of `~/.gitconfig` will be read
from `~/.local/share/chezmoi/home/dot_gitconfig` (instead of
`~/.local/share/chezmoi/dot_gitconfig`).

When migrating an existing chezmoi dotfiles repo to use `.chezmoiroot` you will
need to move the relevant files in to the new root subdirectory manually. You
do not need to move files that are ignored by chezmoi in all cases (i.e. are
listed in `.chezmoiignore` when executed as a template on all machines), and
you can afterwards remove their entries from `home/.chezmoiignore`.

## Use a different version control system to git

Although chezmoi is primarily designed to use a git repo for the source state,
it does not require git and can be used with other version control systems, such
as [fossil](https://www.fossil-scm.org/) or [pijul](https://pijul.org/).

The version control system is used in only three places:

* `chezmoi init` will use `git clone` to clone the source repo if it does not
  already exist.
* `chezmoi update` will use `git pull` by default to pull the latest changes.
* chezmoi's auto add, commit, and push functionality use `git status`, `git
  add`, `git commit` and `git push`.

Using a different version control system (VCS) to git can be achieved in two
ways.

Firstly, if your VCS is compatible with git's CLI, then you can set the
`git.command` configuration variable to your VCS command and set `useBuiltinGit`
to `false`.

Otherwise, you can use your VCS to create the source directory before running
`chezmoi init`, for example:

```sh
fossil clone https://dotfiles.example.com/ dotfiles.fossil
mkdir -p .local/share/chezmoi/.git
cd .local/share/chezmoi
fossil open ~/dotfiles.fossil
chezmoi init --apply
```

!!! note

    The creation of an empty `.git` directory in the source directory is
    required for chezmoi to be able to identify the work tree.

For updates, you can set the `update.command` and `update.args` configuration
variables and `chezmoi update` will use these instead of `git pull`, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[update]
    command = "fossil"
    args = ["update"]
```

Currently, it is not possible to override the auto add, commit, and push
behavior for non-git VCSs, so you will have to commit changes manually, for
example:

```sh
chezmoi cd
fossil add .
fossil commit
```
