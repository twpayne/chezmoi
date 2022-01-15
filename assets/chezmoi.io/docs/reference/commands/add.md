# `add` *target*...

Add *target*s to the source state. If any target is already in the source
state, then its source state is replaced with its current state in the
destination directory.

## `--autotemplate`

Automatically generate a template by replacing strings with variable names from
the `data` section of the config file. Longer substitutions occur before
shorter ones. This implies the `--template` option.

## `-e`, `--empty`

Set the `empty` attribute on added files.

## `--encrypt`

Encrypt files using the defined encryption method.

## `-f`, `--force`

Add *target*s, even if doing so would cause a source template to be
overwritten.

## `--follow`

If the last part of a target is a symlink, add the target of the symlink
instead of the symlink itself.

## `--exact`

Set the `exact` attribute on added directories.

## `-i`, `--include` *types*

Only add entries of type *types*.

## `-p`, `--prompt`

Interactively prompt before adding each file.

## `-r`, `--recursive`

Recursively add all files, directories, and symlinks.

## `-T`, `--template`

Set the `template` attribute on added files and symlinks.

## `--template-symlinks`

When adding symlink to an absolute path in the source directory or destination
directory, create a symlink template with `.chezmoi.sourceDir` or
`.chezmoi.homeDir`. This is useful for creating portable absolute symlinks.

!!! example

    ```console
    $ chezmoi add ~/.bashrc
    $ chezmoi add ~/.gitconfig --template
    $ chezmoi add ~/.ssh/id_rsa --encrypt
    $ chezmoi add ~/.vim --recursive
    $ chezmoi add ~/.oh-my-zsh --exact --recursive
    ```
