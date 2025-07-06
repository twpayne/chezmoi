# Install packages declaratively

chezmoi uses a declarative approach for the contents of dotfiles, but package
installation requires running imperative commands. However, you can simulate
declarative package installation with a combination of a `.chezmoidata` file and
a `run_onchange_` script.

The following example uses [homebrew][homebrew] on macOS, but should be
adaptable to other operating systems and package managers.

First, create `.chezmoidata/packages.yaml` declaring the packages that you want
installed, for example:

```yaml title="~/.local/share/chezmoi/.chezmoidata/packages.yaml"
packages:
  darwin:
    brews:
    - 'git'
    casks:
    - 'google-chrome'
```

Second, create a `run_onchange_darwin-install-packages.sh.tmpl` script that uses
the package manager to install those packages, for example:

``` title="~/.local/share/chezmoi/run_onchange_darwin-install-packages.sh.tmpl"
{{ if eq .chezmoi.os "darwin" -}}
#!/bin/bash

brew bundle --file=/dev/stdin <<EOF
{{ range .packages.darwin.brews -}}
brew {{ . | quote }}
{{ end -}}
{{ range .packages.darwin.casks -}}
cask {{ . | quote }}
{{ end -}}
EOF
{{ end -}}
```

Now, when you run `chezmoi apply`, chezmoi will execute the
`install-packages.sh` script when the list of packages defined in
`.chezmoidata/packages.yaml` changes.

[homebrew]: https://brew.sh
