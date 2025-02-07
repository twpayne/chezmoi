# gopass functions

The `gopass*` template functions return data stored in [gopass][gopass] using
the gopass CLI (`gopass`) or builtin code.

By default, chezmoi will use the gopass CLI (`gopass`). Depending on your gopass
configuration, you may have to enter your passphrase once for each secret.

When setting `gopass.mode` to `builtin`, chezmoi use builtin code to access the
goapass database and caches your passphrase in plaintext in memory until chezmoi
terminates.

!!! warning

    Using the builtin code is experimental and may be removed.

[gopass]: https://www.gopass.pw/
