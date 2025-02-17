# `promptMultichoice` *prompt* *choices* [*default*]

`promptMultichoice` prompts the user with *prompt* and *choices* and returns the
user's response. *choices* must be a list of strings. If a *default* list is
passed and the user's response is empty then it returns *default*.

!!! example

    ```
    {{- $choices := list "chocolate" "strawberry" "vanilla" "pistachio" -}}
    {{- $icecream := promptMultichoice "What type of ice cream do you like"
        $choices (list "pistachio" "chocolate")
    -}}
    [data]
        icecream = {{- $icecream | toToml -}}
    ```
