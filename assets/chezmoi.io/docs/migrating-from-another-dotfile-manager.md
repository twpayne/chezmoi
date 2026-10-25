# Migrating from another dotfile manager

## Migrate from a dotfile manager that uses symlinks

Many dotfile managers replace dotfiles with symbolic links to files in a common
directory. If you [`chezmoi add`][add-command] such a symlink, chezmoi will add
the symlink, not the file. To assist with migrating from symlink-based systems,
use the `--follow` option to `chezmoi add`, for example:

```sh
chezmoi add --follow ~/.bashrc
```

This will tell [`chezmoi add`][add-command] that the target state of `~/.bashrc`
is the target of the `~/.bashrc` symlink, rather than the symlink itself. When
you run [`chezmoi apply`][apply-command], chezmoi will replace the `~/.bashrc`
symlink with the file contents.

[add-command]: /reference/commands/add.md
[apply-command]: /reference/commands/apply.md

