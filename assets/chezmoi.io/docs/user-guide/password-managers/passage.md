# passage

chezmoi includes support for [passage](https://github.com/FiloSottile/passage) using the
passage CLI.

The first line of the output of `passage show $PASS_NAME` is available as the
`passage` template function, for example:

```
{{ passage "$PASS_NAME" }}
```
