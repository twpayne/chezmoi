# Containers and VMs

You can use chezmoi to manage your dotfiles in [GitHub Codespaces][ghc], [Visual
Studio Codespaces][vsc], and [Visual Studio Code Remote - Containers][vscrc].

For a quick start, you can clone the [`chezmoi/dotfiles` repository][czdot]
which supports Codespaces out of the box.

The workflow is different to using chezmoi on a new machine, notably:

* These systems will automatically clone your `dotfiles` repo to `~/dotfiles`,
  so there is no need to clone your repo yourself.

* The installation script must be non-interactive.

* When running in a Codespace, the environment variable `CODESPACES` will be
  set to `true`. You can read its value with the [`env` template
  function][sprig-os].

First, if you are using a chezmoi configuration file template, ensure that it
is non-interactive when running in Codespaces, for example,
`.chezmoi.toml.tmpl` might contain:

```text
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

!!! info

    Setting the `sourceDir` configuration variable to `.chezmoi.sourceDir` is
    required because Codespaces clones your dotfiles repo to a different one to
    chezmoi's default.

This sets the `codespaces` template variable, so you don't have to repeat `(env
"CODESPACES")` in your templates. It also sets the `sourceDir` configuration to
the `--source` argument passed in `chezmoi init`.

Second, create an `install.sh` script that installs chezmoi and your dotfiles
and add it to `.chezmoiignore` and your dotfiles repo:

```sh
chezmoi generate install.sh > install.sh
chmod a+x install.sh
echo install.sh >> .chezmoiignore
git add install.sh .chezmoiignore
git commit -m "Add install.sh"
```

The generated script installs the latest version of chezmoi in `~/.local/bin` if
needed, and then `chezmoi init ...` invokes chezmoi to create its configuration
file and initialize your dotfiles. `--apply` tells chezmoi to apply the changes
immediately, and `--source=...` tells chezmoi where to find the cloned
`dotfiles` repo, which in this case is the same folder in which the script is
running from.

Finally, modify any of your templates to use the `codespaces` variable if
needed. For example, to install `vim-gtk` on Linux but not in Codespaces, your
`run_onchange_install-packages.sh.tmpl` might contain:

```text
{{- if (and (eq .chezmoi.os "linux") (not .codespaces)) -}}
#!/bin/sh
sudo apt install -y vim-gtk
{{- end -}}
```

[ghc]: https://docs.github.com/en/github/developing-online-with-codespaces/personalizing-codespaces-for-your-account
[vsc]: https://code.visualstudio.com/docs/remote/codespaces
[vscrc]: https://code.visualstudio.com/docs/remote/containers#_personalizing-with-dotfile-repositories
[czdot]: https://github.com/chezmoi/dotfiles
[sprig-os]: http://masterminds.github.io/sprig/os.html
