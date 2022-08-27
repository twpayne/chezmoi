# `secret` [*arg*...]

`secret` returns the output of the generic secret command defined by the
`secret.command` configuration variable with `secret.args` and *arg*s with
leading and trailing whitespace removed. The output is cached so multiple calls
to `secret` with the same *arg*s will only invoke the generic secret command
once.
