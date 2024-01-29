# `promptStringOnce` *map* *path* *prompt* [*default*]

`promptStringOnce` returns the value of *map* at *path* if it exists and is a
string value, otherwise it prompts the user for a string value with *prompt* and
an optional *default* using `promptString`.

!!! example

    ```
    {{ $email := promptStringOnce . "email" "What is your email address" }}
    ```
