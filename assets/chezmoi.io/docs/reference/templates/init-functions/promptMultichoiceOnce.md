# `promptMultichoiceOnce` *map* *path* *prompt* *choices* [*default*]

`promptMultichoiceOnce` returns the value of *map* at *path* if it exists and is
a string, otherwise it prompts the user for one of *choices* with *prompt* and
an optional list *default* using [`promptMultichoice`][pm].

!!! example

    ```
    {{- $choices := list "chocolate" "strawberry" "vanilla" "pistachio" -}}
    {{- $icecream := promptMultichoiceOnce
        . "icecream"
        "What type of ice cream do you like"
        $choices
        (list "pistachio" "chocolate")
    -}}
    [data]
        icecream = {{- $icecream | toToml -}}
    ```

[pm]: /reference/templates/init-functions/promptMultichoice.md
