# `shellQuoteList` *list*

`shellQuoteList` returns a list where each element is the corresponding element in
*list* quoted using [`shellQuote`][shellQuote].

!!! example

    ```
    #!/bin/sh

    {{ $args := list "" "$(exec something)" }}
    command {{ $args | shellQuoteList | join " " }}
    ```

[shellQuote]: shellQuote.md
