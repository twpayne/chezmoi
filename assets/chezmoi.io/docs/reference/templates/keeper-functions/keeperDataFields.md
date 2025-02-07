# `keeperDataFields` *uid*

`keeperDataFields` returns the `.data.fields` elements of `keeper get
--format=json *uid*` indexed by `type`.

## Examples

```text
url = {{ (keeperDataFields "$UID").url }}
login = {{ index (keeperDataFields "$UID").login 0 }}
password = {{ index (keeperDataFields "$UID").password 0 }}
```

+++ 2.16.0
