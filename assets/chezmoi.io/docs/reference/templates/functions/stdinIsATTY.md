# `stdinIsATTY`

`stdinIsATTY` returns `true` if chezmoi's standard input is a TTY. It is
primarily useful for determining whether `prompt*` functions should be called in
a config file template or if a default value should be used instead.

!!! example

    ```
    {{ $email := "" }}
    {{ if stdinIsATTY }}
    {{   $email = promptString "email" }}
    {{ else }}
    {{   $email = "user@example.com" }}
    {{ end }}
    ```
