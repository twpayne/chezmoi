# macOS

## Use `brew bundle` to manage your brews and casks

Homebrew's [`brew bundle`
subcommand](https://docs.brew.sh/Manpage#bundle-subcommand) allows you to
specify a list of brews and casks to be installed. You can integrate this with
chezmoi by creating a `run_once_` script. For example, create a file in your
source directory called `run_once_before_install-packages-darwin.sh.tmpl`
containing:

```
{{- if eq .chezmoi.os "darwin" -}}
#!/bin/bash

brew bundle --no-lock --file=/dev/stdin <<EOF
brew "git"
cask "google-chrome"
EOF
{{ end -}}
```

!!! note

    The `Brewfile` is embedded directly in the script with a bash here
    document. chezmoi will run this script whenever its contents change, i.e.
    when you add or remove brews or casks.

## Determine the hostname

The result of the `hostname` command on macOS depends on the network that the
machine is connected to. For a stable result, use the `scutil` command:

```
{{ $computerName := output "scutil" "--get" "ComputerName" | trim }}
```
