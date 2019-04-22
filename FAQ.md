# `chezmoi` FAQ

## What are the consequences of "bare" modifications to the target files?  If my .zshrc is managed by chezmoi and I edit ~/.zshrc without using "chezmoi edit", what happens?

`chezmoi` will overwrite the file the next time you run `chezmoi apply`. Until you run `chezmoi apply` your modified `~/.zshrc` will remain in place.

## How can I tell what dotfiles in my home directory aren't managed by chezmoi? Is there an easy way to have  chezmoi manage a subset of them?

`chezmoi unmanaged` will list everything not managed by `chezmoi`. You can add entire directories with `chezmoi add -r`.

## If there's a mechanism in place for (2) above, is there also a way to tell chezmoi to ignore specific files or groups of files (e.g. by directory name or by glob)?

By default, `chezmoi` ignores everything that you haven't explicitly `chezmoi add`ed. If have files in your source directory that you don't want added to your destination directory when you run `chezmoi apply` add them to a `.chezmoiignore` file (which supports globs and is also a template).

## If the target already exists, but is "behind" the source, can chezmoi be configured to preserve the target version before replacing it with one derived from the source?

Yes. Run `chezmoi add` will update the source state with the target. To see diffs of what would change, without actually changing anything, use `chezmoi diff`.