# `add` *target*...

Add *target*s to the source state. If any target is already in the source
state, then its source state is replaced with its current state in the
destination directory.

## Flags

### `--autotemplate`

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

### `-s`, `--secrets` `ignore`|`warning`|`error`

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

Exclude entries of type [*types*](../command-line-flags/common.md#available-types),
defaults to `none`.

### `-f`, `--force`

Add *target*s, even if doing so would cause a source template to be
overwritten.

### `-i`, `--include` *types*

Only add entries of type [*types*](../command-line-flags/common.md#available-types),
defaults to `all`.

### `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

## Examples

```console
$ chezmoi add ~/.bashrc
$ chezmoi add ~/.gitconfig --template
$ chezmoi add ~/.ssh/id_rsa --encrypt
$ chezmoi add ~/.vim --recursive
$ chezmoi add ~/.oh-my-zsh --exact --recursive
```

## Notes

!!! bug

    `chezmoi add` will fail if the entry being added is in a directory
    implicitly created by an
    [external](../special-files-and-directories/chezmoiexternal-format.md).
    See [this GitHub issue](https://github.com/twpayne/chezmoi/issues/1574) for
    details.
