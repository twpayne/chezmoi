# Dashlane

chezmoi includes support for [Dashlane](https://dashlane.com).

Structured data can be retrieved with the `dashlanePassword` template function,
for example:

```text
examplePassword = {{ (index (dashlanePassword "filter") 0).password }}
```

Secure notes can be retrieved with the `dashlaneNote` template function,
for example:

```text
exampleNote = {{ dashlaneNote "filter" }}
```
