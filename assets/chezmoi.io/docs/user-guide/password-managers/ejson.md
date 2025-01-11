# ejson

chezmoi includes support for [ejson](https://github.com/Shopify/ejson).

Structured data can be retrieved with the `ejsonDecrypt` template function, for
example:

```text
examplePassword = {{ (ejsonDecrypt "my-secrets.ejson").password }}
```

If you want to specify the private key to use for the decryption, structured
data can be retrieved with the `ejsonDecryptWithKey` template function, for
example:

```text
examplePassword = {{ (ejsonDecryptWithKey "my-secrets.ejson" "top-secret-key").password }}
```
