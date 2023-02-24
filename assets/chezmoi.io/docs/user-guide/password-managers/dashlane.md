# Dashlane

chezmoi includes support for [Dashlane](https://dashlane.com).

Structured data can be retrieved with the `dashlanePassword` template function,
for example:

```
examplePassword = {{ (index (dashlanePassword "filter") 0).password }}
```
