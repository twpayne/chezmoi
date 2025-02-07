# `lastpass` *id*

`lastpass` returns structured data from [LastPass][lastpass] using the [LastPass
CLI][cli] (`lpass`). *id* is passed to `lpass show --json $ID` and the output
from `lpass` is parsed as JSON. In addition, the `note` field, if present, is
further parsed as colon-separated key-value pairs. The structured data is an
array so typically the `index` function is used to extract the first item. The
output from `lastpass` is cached so calling `lastpass` multiple times with the
same *id* will only invoke `lpass` once.

!!! example

    ```
    githubPassword = {{ (index (lastpass "GitHub") 0).password | quote }}
    {{ (index (lastpass "SSH") 0).note.privateKey }}
    ```

[lastpass]: https://lastpass.com/
[cli]: https://lastpass.github.io/lastpass-cli/lpass.1.html
