# `bitwardenSecrets` *secret-id* [*access-token*]

`bitwardenSecrets` returns structured data from [Bitwarden][bitwarden] using the
[Bitwarden Secrets CLI][secrets] (`bws`). *secret-id* is passed to `bws secret
get` and the output from `bws secret get` is parsed as JSON and returned.

If the additional *access-token* argument is given, it is passed to `bws secret
get` with the `--access-token` flag.

The output from `bws secret get` is cached so calling `bitwardenSecrets`
multiple times with the same *secret-id* and *access-token* will only invoke
`bws secret get` once.

!!!

    ```
    {{ (bitwardenSecrets "be8e0ad8-d545-4017-a55a-b02f014d4158").value }}
    ```

[bitwarden]: https://bitwarden.com
[secrets]: https://bitwarden.com/help/secrets-manager-cli/
