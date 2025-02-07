# pinentry

By default, chezmoi will request passwords from the terminal.

If the `--no-tty` option is passed, then chezmoi will instead read passwords
from the standard input.

Otherwise, if the configuration variable `pinentry.command` is set then chezmoi
will instead used the given command to read passwords, assuming that it follows
the [Assuan protocol (PDF)][assuan] like [GnuPG's pinentry][pinentry]. The
configuration variable `pinentry.args` specifies extra arguments to be passed to
`pinentry.command` and the configuration variable `pinentry.options` specifies
extra options to be set. The default `pinentry.options` is
`["allow-external-password-cache"]`.

!!! example

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [pinentry]
        command = "pinentry"
    ```

[assuan]: https://www.gnupg.org/documentation/manuals/assuan.pdf
[pinentry]: https://www.gnupg.org/related_software/pinentry/index.html
