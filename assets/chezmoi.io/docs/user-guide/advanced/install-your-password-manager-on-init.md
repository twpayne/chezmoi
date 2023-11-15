# Install your password manager on init

!!! warning

    The approach described here is experimental and may change in a future version
    of chezmoi. If you use this, please contribute to [the
    discussion](https://github.com/twpayne/chezmoi/discussions/3342).

If you use a password manager to store your secrets then you may need to install
your password manager after you have run `chezmoi init` on a new machine but
before `chezmoi init --apply` or `chezmoi apply` executes your `run_before_`
scripts.

chezmoi provides a `hooks.read-source-state.pre` hook that allows you to modify your
system after `chezmoi init` has cloned your dotfile repo but before chezmoi has
read the source state. This is the perfect time to install your password manager
as you can assume that `~/.local/share/chezmoi` is populated but has not yet
been read.

First, write your password manager install hook. chezmoi executes this hook
every time any command reads the source state so the hook should terminate as
quickly as possible if there is no work to do.

This hook is not a template so you cannot use template variables and must
instead detect the system you are running on yourself.

For example:

```sh title="~/.local/share/chezmoi/.install-password-manager.sh"
#!/bin/sh

# exit immediately if password-manager-binary is already in $PATH
type password-manager-binary >/dev/null 2>&1 && exit

case "$(uname -s)" in
Darwin)
    # commands to install password-manager-binary on Darwin
    ;;
Linux)
    # commands to install password-manager-binary on Linux
    ;;
*)
    echo "unsupported OS"
    exit 1
    ;;
esac
```

!!! note

    The leading `.` in `.install-password-manager.sh` is important because it tells
    chezmoi to ignore `.install-password-manager.sh` when declaring the state of
    files in your home directory.

Finally, tell chezmoi to run your password manager install hook before reading
the source state:

```toml title=".config/chezmoi/chezmoi.toml"
[hooks.read-source-state.pre]
    command = ".local/share/chezmoi/.install-password-manager.sh"
```
