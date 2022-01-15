# `fromYaml` *yamltext*

`fromYaml` returns the parsed value of *yamltext*.

!!! example

    ```
    {{ (fromYaml "key1: value\nkey2: value").key2 }}
    ```
