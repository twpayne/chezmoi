# Proton Pass

chezmoi includes support for [Proton Pass][protonpass] using the [Proton Pass
CLI][cli].

Log in to Proton Pass using
```shell
pass-cli login
```

The  output of `pass-cli item view pass://$SHARE_ID/$ITEM_ID/$FIELD` is
available with the [`protonPass`][protonpasstemplatefunc] template function, for
example:

```text
{{ protonPass "pass://$SHARE_ID/$ITEM_ID/$FIELD" }}
```

The  output of `pass-cli item view --output=json pass://$SHARE_ID/$ITEM_ID` is
available with [`protonPassJSON`][protonpassjson] which returns the structured
data the item holds. For example:

```text
{{ (protonPassJSON "pass://$SHARE_ID/$ITEM_ID").item.content.content.key.password }}
```

The contents of attachments are available using the
[`protonPassAttachment`][protonpassattachment] template function, for example:

```
{{ protonPassAttachment "$SHARE_ID" "$ITEM_ID" "$ATTACHMENT_ID" }}
```

[protonpass]: https://proton.me/pass
[protonpassattachment]: ../../reference/templates/protonpass-functions/protonPassAttachment.md
[protonpassjson]: ../../reference/templates/protonpass-functions/protonPassJSON.md
[protonpasstemplatefunc]: ../../reference/templates/protonpass-functions/protonPass.md
[cli]: https://protonpass.github.io/pass-cli
