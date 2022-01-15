# pass

chezmoi includes support for [pass](https://www.passwordstore.org/) using the
pass CLI.

The first line of the output of `pass show <pass-name>` is available as the
`pass` template function, for example:

```
{{ pass "<pass-name>" }}
```
