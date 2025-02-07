# `jq` *query* *input*

`jq` runs the [jq][jq] query *query* against *input* and returns a list of
results.

!!! example

    ```
    {{ dict "key" "value" | jq ".key" | first }}
    ```

!!! warning

    `jq` uses [`github.com/itchyny/gojq`][gojq], which behaves slightly
    differently to the `jq` command in some [edge cases][cases].

[jq]: https://jqlang.org
[gojq]: https://github.com/itchyny/gojq
[cases]: https://github.com/itchyny/gojq#difference-to-jq
