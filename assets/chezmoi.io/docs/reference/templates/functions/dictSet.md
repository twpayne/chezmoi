# `dictSet` *key*... *value* *dict*

`dictSet` modifies *dict* to set *key* to *value* and returns *dict*. Multiple
*key*s may be specified, in which case nested values are set. It is an
alternative to [sprig's `set`
function](http://masterminds.github.io/sprig/dicts.html) with a different
argument order that supports pipelining.

!!! example

    ```
    {{ dict | dictSet "key1" "value1" | dictSet "key2" "nestedKey" "value2" | toJson }}
    ```
