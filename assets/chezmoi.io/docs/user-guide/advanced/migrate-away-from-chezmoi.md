# Migrate away from chezmoi

chezmoi provides several mechanisms to help you move to an alternative dotfile
manager (or even no dotfile manager at all) in the future:

chezmoi creates your dotfiles just as if you were not using a dotfile manager at
all. Your dotfiles are regular files, directories, and symlinks. You can run
[`chezmoi purge`][purge-command] to delete all traces of chezmoi and then, if you're
migrating to a new dotfile manager, then you can use whatever mechanism it
provides to add your dotfiles to your new system.

chezmoi has a [`chezmoi archive`][archive-command] command that generates a tarball of
your dotfiles. You can replace the contents of your dotfiles repo with the
contents of the archive and you've effectively immediately migrated away from
chezmoi.

chezmoi has a [`chezmoi dump`][dump-command] command that dumps the interpreted (target)
state in a machine-readable form, so you can write scripts around chezmoi.

[archive-command]: /reference/commands/archive.md
[dump-command]: /reference/commands/dump.md
[purge-command]: /reference/commands/purge.md
