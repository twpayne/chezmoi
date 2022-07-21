# `promptBoolOnce` *map* *key* *prompt* [*default*]

`promptBoolOnce` returns *map*.*key* if it exists and is a boolean value,
otherwise it prompts the user for a boolean value with *prompt* and an optional
*default* using `promptBool`.

!!! example

    ```
    {{ $hasGUI := promptBoolOnce . "hasGUI" "Does this machine have a GUI" }}
    ```
