# passhole *path* *field*

`passhole` returns the *field* of *path* from a [KeePass][keepass] database
using [passhole][passhole]'s `ph` command.

!!! example

    ```
    {{ passhole "example.com" "password" }}
    ```

[keepass]: https://keepass.info/
[passhole]: https://github.com/Evidlo/passhole
