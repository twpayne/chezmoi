# `promptChoice` *prompt* *choices* [*default*]

`promptChoice` prompts the user with *prompt* and *choices* and returns the user's response. *choices* must be a list of strings. If *default* is passed and the user's response is empty then it returns *default*.

!!! example

    ```
    {{- $choices := list "desktop" "server" -}}
    {{- $hosttype := promptChoice "What type of host are you on" $choices -}}
    [data]
        hosttype = {{- $hosttype | quote -}}
    ```
