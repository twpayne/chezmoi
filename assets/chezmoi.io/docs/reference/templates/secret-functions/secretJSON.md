# `secretJSON` [*arg*...]

`secretJSON` returns structured data from the generic secret command defined by
the `secret.command` configuration variable with `secret.args` and *arg*s. The
output is parsed as JSON. The output is cached so multiple calls to `secret`
with the same *args* will only invoke the generic secret command once.
