# Merge

## Use a custom merge command

By default, chezmoi uses `vimdiff`. You can use a custom command by setting the
`merge.command` and `merge.args` configuration variables. The elements of
`merge.args` are interpreted as templates with the variables `.Destination`,
`.Source`, and `.Target` containing filenames of the file in the destination
state, source state, and target state respectively. For example, to use
[neovim's diff mode][nvim], specify:

```toml title="~/.config/chezmoi/chezmoi.toml"
[merge]
    command = "nvim"
    args = ["-d", "{{ .Destination }}", "{{ .Source }}", "{{ .Target }}"]
```

!!! hint

    If you generate your config file from a config file template, then you'll
    need to escape the `{{` and `}}`. That way your generated config file
    contains the `{{` and `}}` you expect.

    ```toml title="~/.local/share/chezmoi/chezmoi.toml.tmpl"
    [merge]
        command = "nvim"
        args = [
            "-d",
            {{ printf "%q" "{{ .Destination }}" }},
            {{ printf "%q" "{{ .Source }}" }},
            {{ printf "%q" "{{ .Target }}" }},
        ]
    ```

## Use Beyond Compare as the merge tool

To use [Beyond Compare][bcomp] as the merge tool, add the following to your config:

=== "TOML"

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [merge]
        command = "bcomp"
        args = ["{{ .Destination }}", "{{ .Source }}", "{{ .Target }}", "{{ .Source }}"]
    ```

=== "YAML"

    ```yaml title="~/.config/chezmoi/chezmoi.yaml"
    merge:
      command: "bcomp"
      args:
      - "{{ .Destination }}"
      - "{{ .Source }}"
      - "{{ .Target }}"
      - "{{ .Source }}"
    ```

## Use VSCode as the merge tool

To use [VSCode][vscode] as the merge tool, add the following to your config:

=== "TOML"

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [merge]
    command = "bash"
    args = [
        "-c",
        "cp {{ .Target }} {{ .Target }}.base && code --new-window --wait --merge {{ .Destination }} {{ .Target }} {{ .Target }}.base {{ .Source }}",
    ]
    ```

=== "YAML"

    ```yaml title="~/.config/chezmoi/chezmoi.yaml"
    merge:
      command: "bash"
      args:
      - "-c"
      - "cp {{ .Target }} {{ .Target }}.base && code --new-window --wait --merge {{ .Destination }} {{ .Target }} {{ .Target }}.base {{ .Source }}"
    ```

[bcomp]: https://www.scootersoftware.com/
[nvim]: https://neovim.io/doc/user/diff.html
[vscode]: https://code.visualstudio.com/
