# `fromIni` *initext*

`fromIni` returns the parsed value of *initext*.

!!! example

    ```
    {{ (fromIni "[section]\nkey = value").section.key }}
    ```

+++ 2.21.0
