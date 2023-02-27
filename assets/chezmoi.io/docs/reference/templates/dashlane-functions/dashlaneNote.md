# `dashlaneNote` *filter*

`dashlaneNote` returns the content of a secure note from [Dashlane](https://dashlane.com)
using the [Dashlane CLI](https://github.com/Dashlane/dashlane-cli) (`dcli`).
*filter* is passed to `dcli note`, and the output from `dcli
note` is just read as a multiline string.

The output from `dcli note` is cached so calling `dashlaneNote` multiple
times with the same *filter* will only invoke `dcli note` once.

!!! example

    ```
    {{ dashlaneNote "filter" }}
    ```
