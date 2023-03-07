# Hooks

chezmoi runs hooks before and after each chezmoi command as specified by the
`hooks` configuration variable.

Before running *command*, chezmoi runs `hooks.`*command*`.pre.command` with the
arguments `hooks.`*command*`.pre.args`. Similarly, after running *command*,
chezmoi runs `hooks.`*command*`.post.command` with
`hooks.`*command*`.post.args`.

When running hooks, the `CHEZMOI=1` and `CHEZMOI_*` environment variables will
be set. Notably, `CHEZMOI_COMMAND` is set to the chezmoi command being run and
`CHEZMOI_ARGS` contains the full arguments to chezmoi, starting with the path to
chezmoi's executable.

Unlike scripts, hooks are always run, irrespective of whether `--dry-run` is
specified or not.

!!! example

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [hooks.add.pre]
    command = "echo"
    args = ["pre-add-hook"]

    [hooks.apply.post]
    command = "echo"
    args = ["post-apply-hook"]
    ```
