# `add` *target*...

Add *target*s to the source state. If any target is already in the source
state, then its source state is replaced with its current state in the
destination directory.

## `--autotemplate`

Automatically generate a template by replacing strings that match variable
values from the `data` section of the config file with their respective config
names as a template string. Longer substitutions occur before shorter ones.
This implies the `--template` option.

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

## `-q`, `--quiet`

Suppress warnings about adding ignored entries.

## `-r`, `--recursive`

Recursively add all files, directories, and symlinks.

## `-T`, `--template`

Set the `template` attribute on added files and symlinks.

## `--template-symlinks`

When adding symlink to an absolute path in the source directory or destination
directory, create a symlink template with `.chezmoi.sourceDir` or
`.chezmoi.homeDir`. This is useful for creating portable absolute symlinks.

!!! bug

    `chezmoi add` will fail if the entry being added is in a directory
    implicitly created by an
    [external](/reference/special-files-and-directories/chezmoiexternal-format/).
    See [this GitHub issue](https://github.com/twpayne/chezmoi/issues/1574) for
    details.

!!! example

    ```console
    $ chezmoi add ~/.bashrc
    $ chezmoi add ~/.gitconfig --template
    $ chezmoi add ~/.ssh/id_rsa --encrypt
    $ chezmoi add ~/.vim --recursive
    $ chezmoi add ~/.oh-my-zsh --exact --recursive
    ```
