# `isExecutable` *file*

`isExecutable` returns true if a file is executable.

!!! example

    ```
    {{ if isExecutable "/bin/echo" }}
    # echo is executable
    {{ end }}
    ```
