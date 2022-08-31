# passhole *path* *field*

`passhole` returns the *field* of *path* from a [Keepass](https://keypass.info/)
database using [passhole](https://github.com/Evidlo/passhole)'s `ph` command.

!!! example

    ```
    {{ passhole "example.com" "password" }}
    ```
