# `stdinIsATTY`

`stdinIsATTY` returns `true` if chezmoi's standard input is a TTY. It is
primarily useful for determining whether `prompt*` functions should be called
or default values be used.

!!! example

    ```
    {{ $email := "" }}
    {{ if stdinIsATTY }}
    {{   $email = promptString "email" }}
    {{ else }}
    {{   $email = "user@example.com" }}
    {{ end }}
    ```
