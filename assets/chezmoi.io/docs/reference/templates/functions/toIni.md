# `toIni` *value*

`toIni` returns the ini representation of *value*, which must be a dict.

!!! example

    ```
    {{ dict "key" "value" "section" (dict "subkey" "subvalue") | toIni }}
    ```

!!! warning

    The ini format is not well defined, and the particular variant generated
    by `toIni` might not be suitable for you.

+++ 2.21.0
