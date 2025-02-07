# gopass

chezmoi includes support for [gopass][gopass] using the `gopass` CLI.

The first line of the output of `gopass show $PASS_NAME` is available as the
`gopass` template function, for example:

```text
{{ gopass "$PASS_NAME" }}
```

[gopass]: https://www.gopass.pw/
