# `lookOnePath` *file-list*

`lookOnePath` searches for an executable from *file-list* in the directories
named by the `PATH` environment variable. If file contains a slash, it is tried
directly and the `PATH` is not consulted. The result may be an absolute path or
a path relative to the current directory. If an executable from *file-list* is
not found, `lookOnePath` returns an empty string.

`lookOnePath` is not hermetic: its return value depends on the state of the
environment and the filesystem at the moment the template is executed. Exercise
caution when using it in your templates.

The return value of the first successful call to `lookOnePath` is cached, and
future calls to `lookOnePath` for the same *file-list* will return this
path.

`lookOnePath` is provided as an alternative to
[`lookPath`](/reference/templates/functions/lookPath) so that you can look for
one of several variant executables.

!!! example

    ```
    {{ if $ls_alt := lookOnePath (list "eza" "exa") }}
    echo {{ $ls_alt }} is in $PATH
    {{ end }}
    ```

