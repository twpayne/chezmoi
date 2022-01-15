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
