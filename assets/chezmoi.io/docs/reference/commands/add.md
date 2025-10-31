# `add` *target*...

Add *target*s to the source state. If any target is already in the source
state, then its source state is replaced with its current state in the
destination directory.

## Flags

### `-a`, `--autotemplate`

Automatically generate a template by replacing strings that match variable
values from the `data` section of the config file with their respective config
names as a template string. Longer substitutions occur before shorter ones.
This implies the `--template` option.

!!! warning

    `--autotemplate` uses a greedy algorithm which occasionally generates
    templates with unwanted variable substitutions. Carefully review any
    templates it generates.

### `--create`

Add files that should exist, irrespective of their contents.

Sets the `create_` source state attribute on the added file.

A file will be created with the given contents if the file does not exist.
If the file already exists, then its contents will not be changed.
This allows for managing files with an initial state but should not be changed
by chezmoi afterwards.

### `--encrypt`

> Configuration: `add.encrypt`

Encrypt files using the defined encryption method.

### `--exact`

Set the `exact` attribute on added directories.

### `--follow`

If the last part of a target is a symlink, add the target of the symlink
instead of the symlink itself.

### `-p`, `--prompt`

Interactively prompt before adding each file.

### `-q`, `--quiet`

Suppress warnings about adding ignored entries.

### `--secrets` `ignore`|`warning`|`error`

> Configuration: `add.secrets`

Action to take when a secret is found when adding a file. The default is
`warning`.

### `-T`, `--template`

Set the `template` attribute on added files and symlinks.

### `--template-symlinks`

> Configuration: `add.templateSymlinks`

When adding symlink to an absolute path in the source directory or destination
directory, create a symlink template with `.chezmoi.sourceDir` or
`.chezmoi.homeDir`. This is useful for creating portable absolute symlinks.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-f`, `--force`

Add *target*s, even if doing so would cause a source template to be
overwritten.

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-true"

## Examples

```sh
chezmoi add ~/.bashrc
chezmoi add ~/.gitconfig --template
chezmoi add ~/.ssh/id_rsa --encrypt
chezmoi add ~/.vim --recursive
chezmoi add ~/.oh-my-zsh --exact --recursive
```

## Notes

!!! bug

    `chezmoi add` will fail if the entry being added is in a directory
    implicitly created by an [external][external]. See [issue #1574][issue-1574]
    for details.

!!! warning

    `chezmoi add --exact --recursive DIR` works in predictable but surprising
    ways and its use is not recommended for nested directories without taking
    precautions.

    If you have not previously added any files from `~/.config` to chezmoi and
    run `chezmoi add --exact --recursive ~/.config/nvim`, chezmoi wlll consider
    all files under `~/.config` to be managed, and any file *not* in
    `~/.config/nvim` will be removed on your next `chezmoi apply`. This is
    because `~/.config/nvim` is added as:

    ```text
    exact_dot_config/
        exact_nvim/
          exact_lua/
            …
          …
    ```

    To prevent this, add a `.keep` file *first* before adding the subdirectory
    recursively.

    ```sh
    touch ~/.config/.keep
    chezmoi add ~/.config/.keep
    chezmoi add --recursive --exact ~/.config/nvim
    ```

    See [issure #4223][issue-4223] for details.

[external]: /reference/special-files/chezmoiexternal-format.md
[issue-1574]: https://github.com/twpayne/chezmoi/issues/1574
[issue-4223]: https://github.com/twpayne/chezmoi/issues/4223
