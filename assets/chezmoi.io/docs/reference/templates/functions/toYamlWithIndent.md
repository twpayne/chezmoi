# `toYamlWithIndent` *spaces* *value*

`toYamlWithIndent` returns the YAML representation of *value* with an indent of
*spaces* spaces.

!!! example

    ```
    {{ dict "key" (dict "subKey" "value") | toYamlWithIndent 2 }}
    ```
