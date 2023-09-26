# `promptChoiceOnce` *map* *path* *prompt* *choices* [*default*]

`promptChoiceOnce` returns the value of *map* at *path* if it exists and is a
string, otherwise it prompts the user for one of *choices* with *prompt* and an
optional *default* using `promptChoice`.

!!! example

    ```
    {{- $choices := list "desktop" "laptop" "server" "termux" -}}
    {{- $hosttype := promptChoiceOnce . "hosttype" "What type of host are you on" $choices -}}
    [data]
        hosttype = {{- $hosttype | quote -}}
    ```
