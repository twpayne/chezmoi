# Custom

You can use any command line tool that outputs secrets either as a string or in
JSON format. Choose the binary by setting `secret.command` in your
configuration file. You can then invoke this command with the `secret` and
`secretJSON` template functions which return the raw output and JSON-decoded
output respectively. All of the above secret managers can be supported in this
way:

| Secret Manager  | `secret.command` | Template skeleton                                                |
| --------------- | ---------------- | ---------------------------------------------------------------- |
| 1Password       | `op`             | `{{ secretJSON "get" "item" "$ID" }}`                            |
| Bitwarden       | `bw`             | `{{ secretJSON "get" "$ID" }}`                                   |
| Doppler         | `doppler`        | `{{ secretJSON "secrets" "download" "--json" "--no-file" }}`     |
| HashiCorp Vault | `vault`          | `{{ secretJSON "kv" "get" "-format=json" "$ID" }}`               |
| LastPass        | `lpass`          | `{{ secretJSON "show" "--json" "$ID" }}`                         |
| KeePassXC       | `keepassxc-cli`  | Not possible (interactive command only)                          |
| Keeper          | `keeper`         | `{{ secretJSON "get" "--format=json" "$ID" }}`                   |
| pass            | `pass`           | `{{ secret "show" "$ID" }}`                                      |
| passhole        | `ph`             | `{{ secret "$ID" "password" }}`                                  |
