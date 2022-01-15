# `lookPath` *file*

`lookPath` searches for an executable named *file* in the directories named by
the `PATH` environment variable. If file contains a slash, it is tried directly
and the `PATH` is not consulted. The result may be an absolute path or a path
relative to the current directory. If *file* is not found, `lookPath` returns
an empty string.

`lookPath` is not hermetic: its return value depends on the state of the
environment and the filesystem at the moment the template is executed. Exercise
caution when using it in your templates.

!!! example

    ```
    {{ if lookPath "diff-so-fancy" }}
    # diff-so-fancy is in $PATH
    {{ end }}
    ```
