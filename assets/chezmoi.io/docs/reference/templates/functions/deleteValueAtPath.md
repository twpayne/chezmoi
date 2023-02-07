# `deleteValueAtPath` *path* *dict*

`deleteValueAtPath` modifies *dict* to delete the value at *path* and returns
*dict*. *path* can be either a string containing a `.`-separated list of keys or
a list of keys.

If *path* does not exist in *dict* then `deleteValueAtPath` returns *dict*
unchanged.

!!! example

    ```
    {{ dict "outer" (dict "inner" "value") | deleteValueAtPath "outer.inner" | toJson }}
    {{ dict | setValueAtPath "key1" "value1" | setValueAtPath "key2.nestedKey" "value2" | toJson }}
    {{ dict | setValueAtPath (list "key2" "nestedKey") "value2" | toJson }}
    ```
