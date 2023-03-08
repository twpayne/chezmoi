# `jq` *query* *input*

`jq` runs the [jq](https://stedolan.github.io/jq/) query *query* against *input*
and returns a list of results.

!!! example

    ```
    {{ dict "key" "value" | jq ".key" | first }}
    ```

!!! warning

    `jq` uses [`github.com/itchyny/gojq`](https://github.com/itchyny/gojq),
    which behaves slightly differently to the `jq` command in some [edge
    cases](https://github.com/itchyny/gojq#difference-to-jq).
