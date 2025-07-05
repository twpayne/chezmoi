# `comment` *prefix* *text*

`comment` returns *text* with each line prefixed with *prefix*.

`comment` is typically used to comment out blocks of multi-line text
unconditionally. In contrast, the [`ensureLinePrefix` template
function](ensureLinePrefix.md) can be used to only comment out lines that are
not already comments.

!!! example

    ```
    {{ "Line 1\nLine 2\n" | comment "# " }}
    ```
