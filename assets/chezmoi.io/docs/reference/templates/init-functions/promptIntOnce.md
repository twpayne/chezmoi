# `promptIntOnce` *map* *path* *prompt* [*default*]

`promptIntOnce` returns the value of *map* at *path* if it exists and is an
integer value, otherwise it prompts the user for a integer value with *prompt*
and an optional *default* using `promptInt`.

!!! example

    ```
    {{ $monitors := promptIntOnce . "monitors" "How many monitors does this machine have" }}
    ```
