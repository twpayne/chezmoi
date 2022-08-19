# `diff` [*target*...]

Print the difference between the target state and the destination state for
*target*s. If no targets are specified, print the differences for all targets.

If a `diff.pager` command is set in the configuration file then the output will
be piped into it.

If `diff.command` is set then it will be invoked to show individual file
differences with `diff.args` passed as arguments. Each element of `diff.args`
is interpreted as a template with the variables `.Destination` and `.Target`
available corresponding to the path of the file in the source and target state
respectively. The default value of `diff.args` is
`["{{ .Destination }}", "{{ .Target }}"]`. If `diff.args` does not contain any
template arguments then `{{ .Destination }}` and `{{ .Target }}` will be
appended automatically.

## `--reverse`

Reverse the direction of the diff, i.e. show the changes to the target required
to match the destination.

## `--pager` *pager*

Pager to use for output.

## `--use-builtin-diff`

Use chezmoi's builtin diff, even if the `diff.command` configuration variable
is set.

!!! example

    ```console
    $ chezmoi diff
    $ chezmoi diff ~/.bashrc
    ```
