# `quoteList` *list*

`quoteList` returns a list where each element is the corresponding element in
*list* quoted.

!!! example

    ```
    {{ $args := list "alpha" "beta" "gamma" }}
    command {{ $args | quoteList }}
    ```

    ```
    [section]
        array = [{{- $list | quoteList | join ", " -}}]
    ```

+++ 2.18.0
