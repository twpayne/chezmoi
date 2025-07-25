# `abortEmpty`

`abortEmpty` causes template execution to immediately stop and return the empty
string. If `abortEmpty` is called in a sub-template executed by
`includeTemplate` then all template execution stops and returns the empty
string, not just the sub-template.

!!! example

    ```
    {{ abortEmpty }}
    ```
