# `outputList` *name* [*argList*]

`outputList` returns the output of executing the command *name* with the *argList*.
*arg*s. If executing the command returns an error then template execution exits
with an error. The execution occurs every time that the template is executed. It
is the user's responsibility to ensure that executing the command is both
idempotent and fast.

This differs from [`output`](output.md) in that it allows for the *args* to be
created programmatically.

!!! example

    ```
    {{- $args := (list "config" "current-context")  }}
    current-context: {{ outputList "kubectl" $args | trim }}
    ```

