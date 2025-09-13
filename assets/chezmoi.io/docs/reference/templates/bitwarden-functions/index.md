# Bitwarden functions

The `bitwarden*` and `rbw*` functions return data from [Bitwarden][bitwarden]
using the [Bitwarden CLI][bw] (`bw`), [Bitwarden Secrets CLI][bws] (`bws`), and
[`rbw`][rbw] commands.

## Automatic Bitwarden CLI unlock

By default, you must have unlocked your Bitwarden CLI session with the command

```bash
export BW_SESSION="$(bw unlock --raw)"
```

before running chezmoi.

Optionally, you can tell chezmoi to automatically run `bw unlock --raw` and set
the `BW_SESSION` environment variable by setting the `bitwarden.unlock`
configuration variable.  Valid values are:

| Value    | Effect                                                                          |
| -------- | ------------------------------------------------------------------------------- |
| `false`  | Never run `bw unlock --raw` automatically.                                      |
| `true`   | Always run `bw unlock --raw` automatically.                                     |
| `"auto"` | Only run `bw unlock --raw` if the `BW_SESSION` environment variable is not set. |

Additionally, if chezmoi runs `bw unlock raw` automatically, then chezmoi will
also run `bw lock` before terminating.

[bitwarden]: https://bitwarden.com
[bw]: https://bitwarden.com/help/article/cli/
[bws]: https://bitwarden.com/help/secrets-manager-cli/
[rbw]: https://github.com/doy/rbw
