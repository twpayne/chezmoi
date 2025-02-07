# `passFields` *pass-name*

`passFields` returns structured data stored in [pass][pass] using the pass CLI
(`pass`). *pass-name* is passed to `pass show $PASS_NAME` and the output is
parsed as colon-separated key-value pairs, one per line. The return value is
a map of keys to values.

!!! example

    Given the output from `pass`:

    ```
    GitHub
    login: username
    password: secret
    ```

    the return value will be the map:

    ```json
    {
        "login": "username",
        "password": "secret"
    }
    ```

!!! example

    ```
    {{ (passFields "GitHub").password }}
    ```

[pass]: https://www.passwordstore.org
