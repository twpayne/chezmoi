# Passhole
chezmoi includes support for [KeePass](https://keepass.info/) using the
[passhole](https://github.com/Evidlo/passhole) to expose data to a template function.

UserNames and Passwords can be retrieved with the `ph` template function, for
example:
```
exampleUserName = {{ (ph exampleEntry).UserName }}
examplePassword = {{ (ph exampleEntry).Password }}
```

A configuration file can be created at `$HOME/.config/passhole.ini` Based on [passhole's example config](https://github.com/Evidlo/passhole/blob/master/passhole/passhole.ini). For example:
```
[passhole]
default: True
database: ~/example.kdbx
keyfile: ~/example/keyfile
cache: ~/.cache/passhole_cache
cache-timeout: 3600
; no-password: True
```

- database: Path to database (required)
- keyfile: Path to keyfile. If absent, assume no keyfile
- no-password: Whether to prompt for password
- cache: Path to cache password. If absent, don't cache password.
- cache-timeout: Timeout to read from or write to cache while opening this database. No effect if --no-cache=True

Or setting configurations in chezmoi's config file, for example:
```
[passhole]
    command = "ph"
    database = "/example.kdbx"
    keyfile = "/fakekeyfile"
    nocache = true
    nopassword = false
```
Be aware that you can't set *cache-timeout* in this way because it would execute a command like:
`ph --cache-timeout <your timeout seconds> show --field username entry_name`
and there's a bug which would cause the error: "ph: error: argument command: invalid choice: '<your timeout seconds>'".
Because of that I recommend using the `passhole.ini` configuration file.
