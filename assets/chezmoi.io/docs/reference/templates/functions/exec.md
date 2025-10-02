# `exec` *name* [*arg*...]

`exec` executes the command *name* with *arg*s and returns `true` if the command
succeeded, `false` if it failed, or an error if the command cannot be found.
The command's output is ignored. The execution occurs every time that the
template is executed. It is the user's responsibility to ensure that executing
the command is both idempotent and fast.
