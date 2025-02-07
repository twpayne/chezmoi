# `replaceAllRegex` *expr* *repl* *text*

`replaceAllRegex` returns *text* with all substrings matching the regular
expression *expr* replaced with *repl*. It is an alternative to [sprig's
`regexpReplaceAll` function](http://masterminds.github.io/sprig/strings.html)
with a different argument order that supports pipelining.

!!! example

    ```
    {{ "foo subject string" | replaceAllRegex "foo" "bar" }}
    ```

+++ 2.20.0
