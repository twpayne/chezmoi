# `shellQuote` *string*

`shellQuote` returns *string* quoted for POSIX shells.

!!! example

    ```
    #!/bin/sh

    command {{ shellQuote "$untrustedArg" }}
    ```
