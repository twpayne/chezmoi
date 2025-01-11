# Keychain and Windows Credentials Manager

chezmoi includes support for Keychain (on macOS), GNOME Keyring (on Linux and
FreeBSD), and Windows Credentials Manager (on Windows) via the
[`zalando/go-keyring`](https://github.com/zalando/go-keyring) library.

Set values with:

```console
$ chezmoi secret keyring set --service=$SERVICE --user=$USER
Value: xxxxxxxx
```

The value can then be used in templates using the `keyring` function which
takes the service and user as arguments.

For example, save a GitHub access token in keyring with:

```console
$ chezmoi secret keyring set --service=github --user=$GITHUB_USERNAME
Value: xxxxxxxx
```

and then include it in your `~/.gitconfig` file with:

```text
[github]
    user = {{ .github.user | quote }}
    token = {{ keyring "github" .github.user | quote }}
```

You can query the keyring from the command line:

```sh
chezmoi secret keyring get --service=github --user=$GITHUB_USERNAME
```
