# LastPass

chezmoi includes support for [LastPass][lastpass] using the [LastPass CLI][cli]
to expose data as a template function.

Log in to LastPass using:

```sh
lpass login $LASTPASS_USERNAME
```

Check that `lpass` is working correctly by showing password data:

```sh
lpass show --json $LASTPASS_ENTRY_ID
```

where `$LASTPASS_ENTRY_ID` is a [LastPass Entry Specification][spec].

The structured data from `lpass show --json id` is available as the `lastpass`
template function. The value will be an array of objects. You can use the
`index` function and `.Field` syntax of the `text/template` language to extract
the field you want. For example, to extract the `password` field from first the
"GitHub" entry, use:

```text
githubPassword = {{ (index (lastpass "GitHub") 0).password | quote }}
```

chezmoi automatically parses the `note` value of the LastPass entry as
colon-separated key-value pairs, so, for example, you can extract a private SSH
key like this:

```text
{{ (index (lastpass "SSH") 0).note.privateKey }}
```

Keys in the `note` section written as `CamelCase Words` are converted to
`camelCaseWords`.

If the `note` value does not contain colon-separated key-value pairs, then you
can use `lastpassRaw` to get its raw value, for example:

```text
{{ (index (lastpassRaw "SSH Private Key") 0).note }}
```

[lastpass]: https://lastpass.com/
[cli]: https://lastpass.github.io/lastpass-cli/lpass.1.html
[spec]: https://lastpass.github.io/lastpass-cli/lpass.1.html#_entry_specification
