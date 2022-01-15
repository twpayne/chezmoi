# Containers and VMs

You can use chezmoi to manage your dotfiles in [GitHub
Codespaces](https://docs.github.com/en/github/developing-online-with-codespaces/personalizing-codespaces-for-your-account),
[Visual Studio
Codespaces](https://docs.microsoft.com/en/visualstudio/codespaces/reference/personalizing),
and [Visual Studio Code Remote -
Containers](https://code.visualstudio.com/docs/remote/containers#_personalizing-with-dotfile-repositories).

For a quick start, you can clone the [`chezmoi/dotfiles`
repository](https://github.com/chezmoi/dotfiles) which supports Codespaces out
of the box.

The workflow is different to using chezmoi on a new machine, notably:

* These systems will automatically clone your `dotfiles` repo to `~/dotfiles`,
  so there is no need to clone your repo yourself.

* The installation script must be non-interactive.

* When running in a Codespace, the environment variable `CODESPACES` will be
  set to `true`. You can read its value with the [`env` template
  function](http://masterminds.github.io/sprig/os.html).

First, if you are using a chezmoi configuration file template, ensure that it
is non-interactive when running in Codespaces, for example,
`.chezmoi.toml.tmpl` might contain:

```
{{- $codespaces:= env "CODESPACES" | not | not -}}
sourceDir = {{ .chezmoi.sourceDir | quote }}

[data]
    name = "Your name"
    codespaces = {{ $codespaces }}
{{- if $codespaces }}{{/* Codespaces dotfiles setup is non-interactive, so set an email address */}}
    email = "your@email.com"
{{- else }}{{/* Interactive setup, so prompt for an email address */}}
    email = {{ promptString "email" | quote }}
{{- end }}
```

This sets the `codespaces` template variable, so you don't have to repeat `(env
"CODESPACES")` in your templates. It also sets the `sourceDir` configuration to
the `--source` argument passed in `chezmoi init`.

Second, create an `install.sh` script that installs chezmoi and your dotfiles:

```sh
#!/bin/sh

set -e # -e: exit on error

if [ ! "$(command -v chezmoi)" ]; then
  bin_dir="$HOME/.local/bin"
  chezmoi="$bin_dir/chezmoi"
  if [ "$(command -v curl)" ]; then
    sh -c "$(curl -fsLS https://chezmoi.io/get)" -- -b "$bin_dir"
  elif [ "$(command -v wget)" ]; then
    sh -c "$(wget -qO- https://chezmoi.io/get)" -- -b "$bin_dir"
  else
    echo "To install chezmoi, you must have curl or wget installed." >&2
    exit 1
  fi
else
  chezmoi=chezmoi
fi

# POSIX way to get script's dir: https://stackoverflow.com/a/29834779/12156188
script_dir="$(cd -P -- "$(dirname -- "$(command -v -- "$0")")" && pwd -P)"
# exec: replace current process with chezmoi init
exec "$chezmoi" init --apply "--source=$script_dir"
```

Ensure that this file is executable (`chmod a+x install.sh`), and add
`install.sh` to your `.chezmoiignore` file.

It installs the latest version of chezmoi in `~/.local/bin` if needed, and then
`chezmoi init ...` invokes chezmoi to create its configuration file and
initialize your dotfiles. `--apply` tells chezmoi to apply the changes
immediately, and `--source=...` tells chezmoi where to find the cloned
`dotfiles` repo, which in this case is the same folder in which the script is
running from.

If you do not use a chezmoi configuration file template you can use `chezmoi
apply --source=$HOME/dotfiles` instead of `chezmoi init ...` in `install.sh`.

Finally, modify any of your templates to use the `codespaces` variable if
needed. For example, to install `vim-gtk` on Linux but not in Codespaces, your
`run_once_install-packages.sh.tmpl` might contain:

```
{{- if (and (eq .chezmoi.os "linux") (not .codespaces)) -}}
#!/bin/sh
sudo apt install -y vim-gtk
{{- end -}}
```
