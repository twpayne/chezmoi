# passhole *path* *field*

`passhole` returns the *field* of *path* from a [KeePass](https://keepass.info/)
database using [passhole](https://github.com/Evidlo/passhole)'s `ph` command.

!!! example

    ```
    {{ passhole "example.com" "password" }}
    ```

+++ 2.23.0
