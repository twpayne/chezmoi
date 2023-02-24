# `dashlanePassword` *filter*

`dashlanePassword` returns structured data from [Dashlane](https://dashlane.com)
using the [Dashlane CLI](https://github.com/Dashlane/dashlane-cli) (`dcli`).
*filter* is passed to `dcli password --output json`, and the output from `dcli
password` is parsed as JSON.

The output from `dcli password` cached so calling `dashlanePassword` multiple
times with the same *filter* will only invoke `dcli password` once.

!!! example

    ```
    {{ (index (dashlanePassword "filter") 0).password }}
    ```
