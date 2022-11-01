# `setValueAtPath` *path* *value* *dict*

`setValueAtPath` modifies *dict* to set the value at *path* to *value* and
returns *dict*. *path* can be either a string containing a `.`-separated list of
keys or a list of keys. The function will create new key/value pairs in *dict*
if needed.

This is an alternative to [sprig's `set`
function](http://masterminds.github.io/sprig/dicts.html) with a different
argument order that supports pipelining.

!!! example

    ```
    {{ dict | setValueAtPath "key1" "value1" | setValueAtPath "key2.nestedKey" "value2" | toJson }}
    {{ dict | setValueAtPath (list "key2" "nestedKey") "value2" | toJson }}
    ```
