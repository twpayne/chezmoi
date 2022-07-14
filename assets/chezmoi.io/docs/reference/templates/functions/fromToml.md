# `fromToml` *tomltext*

`fromToml` returns the parsed value of *tomltext*.

!!! example

    ```
    {{ (fromToml "[section]\nkey = \"value\"").section.key }}
    ```
