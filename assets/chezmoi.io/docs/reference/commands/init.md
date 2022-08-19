# `init` [*repo*]

Setup the source directory, generate the config file, and optionally update the
destination directory to match the target state.

By default, if *repo* is given, chezmoi will guess the full git repo URL, using
HTTPS by default, or SSH if the `--ssh` option is specified, according to the
following patterns:

| Pattern            | HTTPS Repo                             | SSH repo                           |
| ------------------ | -------------------------------------- | ---------------------------------- |
| `user`             | `https://github.com/user/dotfiles.git` | `git@github.com:user/dotfiles.git` |
| `user/repo`        | `https://github.com/user/repo.git`     | `git@github.com:user/repo.git`     |
| `site/user/repo`   | `https://site/user/repo.git`           | `git@site:user/repo.git`           |
| `~sr.ht/user`      | `https://git.sr.ht/~user/dotfiles`     | `git@git.sr.ht:~user/dotfiles.git` |
| `~sr.ht/user/repo` | `https://git.sr.ht/~user/repo`         | `git@git.sr.ht:~/user/repo.git`    |

To disable git repo URL guessing pass the `--guess-repo-url=false` option.

First, if the source directory is not already contain a repository, then if
*repo* is given it is checked out into the source directory, otherwise a new
repository is initialized in the source directory.

Second, if a file called `.chezmoi.$FORMAT.tmpl` exists, where `$FORMAT` is one
of the supported file formats (e.g. `json`, `toml`, or `yaml`) then a new
configuration file is created using that file as a template.

Then, if the `--apply` flag is passed, `chezmoi apply` is run.

Then, if the `--purge` flag is passed, chezmoi will remove its source, config,
and cache directories.

Finally, if the `--purge-binary` is passed, chezmoi will attempt to remove its
own binary.

## `--apply`

Run `chezmoi apply` after checking out the repo and creating the config file.

## `--branch` *branch*

Check out *branch* instead of the default branch.

## `--config-path` *path*

Write the generated config file to *path* instead of the default location.

## `--data` *bool*

Include existing template data when creating the config file. This defaults to
`true`. Set this to `false` to simulate creating the config file with no
existing template data.

## `--depth` *depth*

Clone the repo with depth *depth*.

## `--prompt`

Force the `prompt*Once` template functions to prompt.

## `--guess-repo-url` *bool*

Guess the repo URL from the *repo* argument. This defaults to `true`.

## `--one-shot`

`--one-shot` is the equivalent of `--apply`, `--depth=1`, `--force`, `--purge`,
and `--purge-binary`. It attempts to install your dotfiles with chezmoi and
then remove all traces of chezmoi from the system. This is useful for setting
up temporary environments (e.g. Docker containers).

## `--purge`

Remove the source and config directories after applying.

## `--purge-binary`

Attempt to remove the chezmoi binary after applying.

## `--ssh`

Guess an SSH repo URL instead of an HTTPS repo.

!!! example

    ```console
    $ chezmoi init user
    $ chezmoi init user --apply
    $ chezmoi init user --apply --purge
    $ chezmoi init user/dots
    $ chezmoi init codeberg.org/user
    $ chezmoi init gitlab.com/user
    ```
