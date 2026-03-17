# pass

chezmoi includes support for [Proton Pass][protonpass] using [proton pass CLI][cli].

Log in to proton pass using 
```shell
pass-cli login
```

The  output of `pass-cli item view pass://$SHARE_ID/$ITEM_ID/$FIELD` is available as the
`protonPass` template function, for example:

```text
{{ protonPass "pass://$SHARE_ID/$ITEM_ID/$FIELD" }}
```

The  output of `pass-cli item view pass://$SHARE_ID/$ITEM_ID` is available as the
`protonPassJSON` and returns the structured data the item holds. For example:

```text
{{ (protonPassJSON "pass://$SHARE_ID/$ITEM_ID").item.content.content.key.password }}
```

[protonpass]: https://proton.me/pass
[cli]: https://protonpass.github.io/pass-cli/
