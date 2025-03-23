# `init` [*repo*]

Setup the source directory, generate the config file, and optionally update the
destination directory to match the target state. This is done in the following
order:

1. The source directory is initialized. If chezmoi does not detect a Git
   repository in the source directory, chezmoi will clone the provided *repo*
   into the source directory. If no *repo* is provided, chezmoi will initialize
   a new Git repository.

2. If the initialized source directory contains a `.chezmoi.$FORMAT.tmpl` file,
   a new configuration file will be created using that file as a template.

3. If the `--apply` flag is provided, `chezmoi apply` is run.

4. If the `--purge` flag is provided, chezmoi will remove the source, config,
   and cache directories.

5. If the `--purge-binary` is passed, chezmoi will attempt to remove its own
   binary.

By default, if *repo* is given, chezmoi will guess the full git repo URL, using
HTTPS by default, or SSH if the `--ssh` option is specified, according to the
following patterns:

| Pattern            | HTTPS Repo                                  | SSH repo                           |
| ------------------ | ------------------------------------------- | ---------------------------------- |
| `user`             | `https://user@github.com/user/dotfiles.git` | `git@github.com:user/dotfiles.git` |
| `user/repo`        | `https://user@github.com/user/repo.git`     | `git@github.com:user/repo.git`     |
| `site/user/repo`   | `https://user@site/user/repo.git`           | `git@site:user/repo.git`           |
| `sr.ht/~user`      | `https://user@git.sr.ht/~user/dotfiles`     | `git@git.sr.ht:~user/dotfiles.git` |
| `sr.ht/~user/repo` | `https://user@git.sr.ht/~user/repo`         | `git@git.sr.ht:~user/repo.git`     |

To disable git repo URL guessing, pass the `--guess-repo-url=false` option.

!!! info

    If you are using a different version control system, there are different
    steps [required for repo initialization][alt-vcs]. To prevent chezmoi from
    trying to clone or create a Git repository, add an empty `.git` directory to
    the source directory.

    ```sh
    mkdir -p ~/.local/share/chezmoi/.git
    ```

--8<-- "config-format.md"

## Flags

### `-a`, `--apply`

Run `chezmoi apply` after checking out the repo and creating the config file.

### `--branch` *branch*

Check out *branch* instead of the default branch.

### `-C`, `--config-path` *path*

Write the generated config file to *path* instead of the default location.

### `--data` *bool*

Include existing template data when creating the config file. This defaults to
`true`. Set this to `false` to simulate creating the config file with no
existing template data.

### `-d`, `--depth` *depth*

Clone the repo with depth *depth*.

### `--git-lfs` *bool*

Run `git lfs pull` after cloning the repo.

### `-g`, `--guess-repo-url` *bool*

Guess the repo URL from the *repo* argument. This defaults to `true`.

### `--one-shot`

`--one-shot` is the equivalent of `--apply`, `--depth=1`, `--force`, `--purge`,
and `--purge-binary`. It attempts to install your dotfiles with chezmoi and then
remove all traces of chezmoi from the system. This is useful for setting up
temporary environments (e.g. Docker containers).

### `--prompt`

Force the `prompt*Once` template functions to prompt.

### `--promptBool` *pairs*

Populate the `promptBool` template function with values from *pairs*. *pairs* is
a comma-separated list of *prompt*`=`*value* pairs. If `promptBool` is called
with a *prompt* that does not match any of *pairs*, then it prompts the user for
a value.

### `--promptChoice` *pairs*

Populate the `promptChoice` template function with values from *pairs*. *pairs*
is a comma-separated list of *prompt*`=`*value* pairs. If `promptChoice` is
called with a *prompt* that does not match any of *pairs*, then it prompts the
user for a value.

### `--promptDefaults`

Make all `prompt*` template function calls with a default value return that
default value instead of prompting.

### `--promptInt` *pairs*

Populate the `promptInt` template function with values from *pairs*. *pairs* is
a comma-separated list of *prompt*`=`*value* pairs. If `promptInt` is called
with a *prompt* that does not match any of *pairs*, then it prompts the user for
a value.

### `--promptMultichoice` *pairs*

Populate the `promptMultichoice` template function with values from *pairs*.
*pairs* is a comma-separated list of *prompt*`=`*value*[`/`*value*] pairs. If
`promptMultichoice` is called with a *prompt* that does not match any of
*pairs*, then it prompts the user for values.

### `--promptString` *pairs*

Populate the `promptString` template function with values from *pairs*. *pairs*
is a comma-separated list of *prompt*`=`*value* pairs. If `promptString` is
called with a *prompt* that does not match any of *pairs*, then it prompts the
user for a value.

### `-p`, `--purge`

Remove the source and config directories after applying.

### `-P`, `--purge-binary`

Attempt to remove the chezmoi binary after applying.

### `--recurse-submodules` *bool*

Recursively clone submodules. This defaults to `true`.

### `--ssh`

Guess an SSH repo URL instead of an HTTPS repo.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

## Examples

```sh
chezmoi init user
chezmoi init user --apply
chezmoi init user --apply --purge
chezmoi init user/dots
chezmoi init codeberg.org/user
chezmoi init gitlab.com/user
```

[alt-vcs]: /user-guide/advanced/customize-your-source-directory.md#use-a-different-version-control-system-to-git
