# `output` *name* [*arg*...]

`output` returns the output of executing the command *name* with *arg*s. If
executing the command returns an error then template execution exits with an
error. The execution occurs every time that the template is executed. It is the
user's responsibility to ensure that executing the command is both idempotent
and fast.

!!! example

    ```
    current-context: {{ output "kubectl" "config" "current-context" | trim }}
    ```
