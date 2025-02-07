# `dashlaneNote` *filter*

`dashlaneNote` returns the content of a secure note from [Dashlane][dashlane]
using the [Dashlane CLI][cli] (`dcli`). *filter* is passed to `dcli note`, and
the output from `dcli note` is just read as a multi-line string.

The output from `dcli note` is cached so calling `dashlaneNote` multiple times
with the same *filter* will only invoke `dcli note` once.

!!! example

    ```
    {{ dashlaneNote "filter" }}
    ```

[dashlane]: https://dashlane.com
[cli]: https://github.com/Dashlane/dashlane-cli
