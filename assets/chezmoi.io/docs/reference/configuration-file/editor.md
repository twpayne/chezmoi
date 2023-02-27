# Editor

The editor used is the first non-empty string of the `edit.command`
configuration variable, the `$VISUAL` environment variable, the `$EDITOR`
environment variable. If none are set then chezmoi falls back to `notepad.exe`
on Windows systems and `vi` on non-Windows systems.

When the `edit.command` configuration variable is used, extra arguments can be
passed to the editor with the `edit.args` configuration variable.

chezmoi will emit a warning if the editor returns in less than
`edit.minDuration` (default `1s`). To disable this warning, set
`edit.minDuration` to `0`.
