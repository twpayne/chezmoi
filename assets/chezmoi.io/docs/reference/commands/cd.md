# `cd` [*path*]

Launch a shell in the working tree (typically the source directory). chezmoi
will launch the command set by the `cd.command` configuration variable with any
extra arguments specified by `cd.args`. If this is not set, chezmoi will
attempt to detect your shell and finally fall back to an OS-specific default.

If the optional argument *path* is present, the shell will be launched in the
source directory corresponding to *path*.

!!! hint

    This does not change the current directory of the current shell. To do
    that, instead use:

    ```console
    $ cd $(chezmoi source-path)
    ```

!!! example

    ```console
    $ chezmoi cd
    ```
