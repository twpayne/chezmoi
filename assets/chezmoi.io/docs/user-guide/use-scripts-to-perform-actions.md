# Use scripts to perform actions

## Understand how scripts work

chezmoi supports scripts that are executed when you run
[`chezmoi apply`](/reference/commands/apply.md). These scripts can be configured
to run every time, only when their contents have changed, or only if they
haven't been run before.

Scripts are any file in the source directory with the prefix `run_`, and
they are executed in alphabetical order.

- **`run_` scripts**: These scripts are executed every time you run `chezmoi apply`.
- **`run_onchange_` scripts**: These scripts are only executed if their content
has changed since the last time they were run.
- **`run_once_` scripts**: These scripts are executed once for each unique
version of the content. If the script is a template, the content is hashed after
template execution. chezmoi tracks the content's SHA256 hash and stores it in
a database. If the content has been run before (even under a different filename),
the script will not run again unless the content itself changes.

Scripts break chezmoi's declarative approach and should be used sparingly.
All scripts should be idempotent, including `run_onchange_` and `run_once_` scripts.

Scripts are normally run while chezmoi updates your dotfiles. For example,
`run_b.sh` will be run after updating `a.txt` and before updating `c.txt`.
To run scripts before or after the updates, use the `before_` or `after_`
attributes, respectively, e.g., `run_once_before_install-password-manager.sh`.

Scripts must be created manually in the source directory, typically by running
[`chezmoi cd`](/reference/commands/cd.md) and then creating a file with a `run_`
prefix. There is no need to set the executable bit on the script.

Scripts with the `.tmpl` suffix are treated as templates, with the usual
template variables available. If the template resolves to only whitespace
or an empty string, the script will not be executed, which is useful for
disabling scripts dynamically.

When executing a script, chezmoi generates the script contents in a file in a
temporary directory with the executable bit set and then executes it using `exec(3)`.
As a result, the script must either include a `#!` line or be an executable binary.
Script working directory is set to the first existing parent directory in the
destination tree.

If a `.chezmoiscripts` directory exists at the root of the source directory,
scripts in this directory are executed as normal scripts, without creating
a corresponding directory in the target state.

In _verbose_ mode, the scripts' contents are printed before execution.
In _dry-run_ mode, scripts are not executed.

## Set environment variables

You can set extra environment variables for your scripts in the `scriptEnv`
section of your config file. For example, to set the `MY_VAR` environment
variable to `my_value`, specify:

```toml title="~/.config/chezmoi/chezmoi.toml"
[scriptEnv]
    MY_VAR = "my_value"
```

chezmoi sets a number of environment variables when running scripts, including
`CHEZMOI=1` and common template data like `CHEZMOI_OS` and `CHEZMOI_ARCH`.

!!! note

    By default, `chezmoi diff` will print the contents of scripts that would be
    run by `chezmoi apply`. To exclude scripts from the output of `chezmoi
    diff`, set `diff.exclude` in your configuration file, for example:

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [diff]
        exclude = ["scripts"]
    ```

    Similarly, `chezmoi status` will print the names of the scripts that it
    will execute with the status `R`. This can similarly disabled by setting
    `status.exclude` to `["scripts"]` in your configuration file.

## Install packages with scripts

Change to the source directory and create a file called
`run_onchange_install-packages.sh`. In this file create your package
installation script, e.g.

```sh
#!/bin/sh
sudo apt install ripgrep
```

The next time you run [`chezmoi apply`](/reference/commands/apply.md) or
[`chezmoi update`](/reference/commands/update.md) this script will be
run. As it has the `run_onchange_` prefix, it will not be run again unless its
contents change, for example if you add more packages to be installed.

This script can also be a template. For example, if you create
`run_onchange_install-packages.sh.tmpl` with the contents:

``` title="~/.local/share/chezmoi/run_onchange_install-packages.sh.tmpl"
{{ if eq .chezmoi.os "linux" -}}
#!/bin/sh
sudo apt install ripgrep
{{ else if eq .chezmoi.os "darwin" -}}
#!/bin/sh
brew install ripgrep
{{ end -}}
```

This will install `ripgrep` on both Debian/Ubuntu Linux systems and macOS.

## Run a script when the contents of another file changes

chezmoi's `run_` scripts are run every time you run
[`chezmoi apply`](/reference/commands/apply.md), whereas
`run_onchange_` scripts are run only when their contents have changed, after
executing them as templates. You can use this to cause a `run_onchange_` script
to run when the contents of another file has changed by including a checksum of
the other file's contents in the script.

For example, if your [dconf](https://wiki.gnome.org/Projects/dconf) settings
are stored in `dconf.ini` in your source directory then you can make `chezmoi
apply` only load them when the contents of `dconf.ini` has changed by adding
the following script as `run_onchange_dconf-load.sh.tmpl`:

``` title="~/.local/share/chezmoi/run_onchange_dconf-load.sh.tmpl"
#!/bin/bash

# dconf.ini hash: {{ include "dconf.ini" | sha256sum }}
dconf load / < {{ joinPath .chezmoi.sourceDir "dconf.ini" | quote }}
```

As the SHA256 sum of `dconf.ini` is included in a comment in the script, the
contents of the script will change whenever the contents of `dconf.ini` are
changed, so chezmoi will re-run the script whenever the contents of `dconf.ini`
change.

In this example you should also add `dconf.ini` to
[`.chezmoiignore`](/reference/special-files/chezmoiignore.md) so chezmoi
does not create `dconf.ini` in your home directory.

## Clear the state of all `run_onchange_` and `run_once_` scripts

chezmoi stores whether and when `run_onchange_` and `run_once_` scripts have
been run in its persistent state.

To clear the state of `run_onchange_` scripts, run:

```sh
chezmoi state delete-bucket --bucket=entryState
```

To clear the state of `run_once_` scripts, run:

```sh
chezmoi state delete-bucket --bucket=scriptState
```
