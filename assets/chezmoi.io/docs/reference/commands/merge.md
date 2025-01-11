# `merge` *target*...

Perform a three-way merge between the destination state, the target state, and
the source state for each *target*. The merge tool is defined by the
`merge.command` configuration variable, and defaults to `vimdiff`. If multiple
targets are specified the merge tool is invoked separately and sequentially for
each target. If the target state cannot be computed (for example if source is a
template containing errors or an encrypted file that cannot be decrypted) a
two-way merge is performed instead.

The order of arguments to `merge.command` is set by `merge.args`. Each argument
is interpreted as a template with the variables `.Destination`, `.Source`, and
`.Target` available corresponding to the path of the file in the destination
state, the source state, and the target state respectively. The default value
of `merge.args` is `["{{ .Destination }}", "{{ .Source }}", "{{ .Target }}"]`.
If `merge.args` does not contain any template arguments then `{{ .Destination
}}`, `{{ .Source }}`, and `{{ .Target }}` will be appended automatically.

## Examples

```sh
chezmoi merge ~/.bashrc
```
