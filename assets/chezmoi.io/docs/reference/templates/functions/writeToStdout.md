# `writeToStdout` *string*...

`writeToStdout` writes each *string* to stdout. It is only available when
generating the initial config file.

!!! example

    ```
    {{- writeToStdout "Hello, world\n" -}}
    ```
