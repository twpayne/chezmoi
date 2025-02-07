# pass

chezmoi includes support for [pass][pass] using the pass CLI.

The first line of the output of `pass show $PASS_NAME` is available as the
`pass` template function, for example:

```text
{{ pass "$PASS_NAME" }}
```

[pass]: https://www.passwordstore.org/
