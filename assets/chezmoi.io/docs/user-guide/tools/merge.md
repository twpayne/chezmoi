# Merge

## Use a custom merge command

By default, chezmoi uses `vimdiff`. You can use a custom command by setting the
`merge.command` and `merge.args` configuration variables. The elements of
`merge.args` are interpreted as templates with the variables `.Destination`,
`.Source`, and `.Target` containing filenames of the file in the destination
state, source state, and target state respectively. For example, to use
[neovim's diff mode](https://neovim.io/doc/user/diff.html), specify:

```toml title="~/.config/chezmoi/chezmoi.toml"
[merge]
    command = "nvim"
    args = ["-d", "{{ .Destination }}", "{{ .Source }}", "{{ .Target }}"]
```

!!! hint

    If you generate your config file from a config file template, then you'll
    need to escape the `{{` and `}}` as `{{"{{"}}` and `{{"}}"}}`. That way 
    your generated config file contains the `{{` and `}}` you expect.
