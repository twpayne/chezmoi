# Troubleshooting

## How can I quickly check for problems with chezmoi on my machine?

Run:

```console
$ chezmoi doctor
```

Anything `ok` is fine, anything `warning` is only a problem if you want to use
the related feature, and anything `error` indicates a definite problem.

## A specific command is not behaving as I expect. How can I debug it?

The `--verbose` flag makes chezmoi to print extra information about what it
is doing.

The `--debug` flag makes chezmoi print very detailed step by step information.

## The output of `chezmoi diff` is broken and does not contain color. What could be wrong?

By default, chezmoi's diff output includes ANSI color escape sequences (e.g.
`ESC[37m`) and is piped into your pager (by default `less`). chezmoi assumes
that your pager passes through the ANSI color escape sequences, as configured on
many systems, but not all. If your pager does not pass through ANSI color escape
sequences then you will see monochrome diff output with uninterpreted ANSI color
escape sequences.

This can typically by fixed by setting the environment variable

```console
$ export LESS=-R
```

which instructs `less` to display "raw" control characters via the `-R` /
`--RAW-CONTROL-CHARS` option.

You can also set the `pager` configuration variable in your config file, for
example:

```toml title="~/.config/chezmoi/chezmoi.toml"
pager = "less -R"
```

If you have set a different pager (via the `pager` configuration variable or
`PAGER` environment variable) then you must ensure that it passes through raw
control characters. Alternatively, you can use the `--color=false` option to
chezmoi to disable colors or the `--no-pager` option to chezmoi to disable the
pager.

## Why do I get a blank buffer or empty file when running `chezmoi edit`?

In this case, `chezmoi edit` typically prints a warning like:

```
chezmoi: warning: $EDITOR $TMPDIR/$FILENAME: returned in less than 1s
```

`chezmoi edit` performs a bit of magic to improve the experience of editing
files in the source state by invoking your editor with filenames in a temporary
directory that look like filenames in your home directory. What's happening
here is that your editor command is exiting immediately, so chezmoi thinks
you've finished editing and so removes the temporary directory, but actually
your editor command has forked a edit process in the background, and that edit
process opens a now non-existent file.

To fix this you have to configure your editor command to remain in the
foreground until you have finished editing the file, so chezmoi knows when to
remove the temporary directory.

=== "VIM"

    Pass the `-f` flag, e.g. by setting the `edit.flags` configuration variable
    to `["-f"]`, or by setting the `EDITOR` environment variable to include the
    `-f` flag, e.g. `export EDITOR="vim -f"`.

=== "VSCode"

    Pass the `--wait` flag, e.g. by setting the `edit.flags` configuration
    variable to `["--wait"]` or by setting the `EDITOR` environment variable to
    include the `--wait` flag, e.g. `export EDITOR="code --wait"`.

=== "Any editor"

    In this example 'mousepad' is the editor to be used. This program will not wait in the foreground by default. It also requires DBUS server integration to be disabled and it needs to start a new window for each execution (otherwise new PIDs will not be created for the process, if another mousepad instance is already running)

    1. Add a new bash script somewhere in your $PATH (like here: `$HOME/.local/bin/chezmoi-mousepad`)

    ```bash
    #!/usr/bin/env bash
    wait -n; mousepad --opening-mode=window --encoding=UTF-8 --disable-server "$@"
    ```

    The key here is the `wait -n` statement that will wait for the next child job to complete and can be universally applied (within the context of a bash script) to any editor you choose.

    2. Run `chmod +x $HOME/.local/bin/chezmoi-mousepad` to make the script executable
    3. Finally add to (or update) `chezmoi.toml`
    
    ```toml
    [edit]
        command = "chezmoi-mousepad"
        args = []
    ```

    or by setting the `EDITOR` environment variable, e.g. `export EDITOR="chezmoi-mousepad"`.

The "bit of magic" that `chezmoi edit` performs includes:

* `chezmoi edit` makes the filename opened by your editor more closely match
  the target filename, which can help your editor choose the correct syntax
  highlighting. For example, if you run `chezmoi edit ~/.zshrc`, your editor is
  be opened with `$TMPDIR/.zshrc` but you'll actually be editing
  `~/.local/share/chezmoi/dot_zshrc`. Under the hood, chezmoi creates a
  hardlink in a temporary directory to the file in your source directory, so
  even though your editor thinks it's editing `.zshrc`, it is really editing
  `dot_zshrc` in your source directory.

* If the source file is encrypted then `chezmoi edit` transparently decrypts
  and re-encrypts the file for you. Specifically, chezmoi decrypts the file
  into a private temporary directory and open your editor with the decrypted
  file, and re-encrypts the file when you exit your editor.

* If the source file is a template, then `chezmoi edit` preserves the `.tmpl`
  extension.

## chezmoi makes `~/.ssh/config` group writeable. How do I stop this?

By default, chezmoi uses your system's umask when creating files. On most
systems the default umask is `022` but some systems use `002`, which means
that files and directories are group writeable by default.

You can override this for chezmoi by setting the `umask` configuration variable
in your configuration file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
umask = 0o022
```

!!! note

    This will apply to all files and directories that chezmoi manages and will
    ensure that none of them are group writeable. It is not currently possible
    to control group write permissions for individual files or directories.
    Please [open an issue on
    GitHub](https://github.com/twpayne/chezmoi/issues/new?assignees=&labels=enhancement&template=02_feature_request.md&title=)
    if you need this.

## chezmoi reports `chezmoi: user: lookup userid NNNNN: input/output error`

This is likely because the chezmoi binary you are using was statically compiled
with [musl](https://musl.libc.org/) and the machine you are running on uses
LDAP or NIS.

The immediate fix is to use a package built for your distribution (e.g a `.deb`
or `.rpm`) which is linked against glibc and includes LDAP/NIS support instead
of the statically-compiled binary.

If the problem still persists, then please [open an issue on
GitHub](https://github.com/twpayne/chezmoi/issues/new/choose).

## chezmoi reports `chezmoi: timeout` or `chezmoi: timeout obtaining persistent state lock`

chezmoi will report this when it is unable to lock its persistent state
(`~/.config/chezmoi/chezmoistate.boltdb`), typically because another instance of
chezmoi is currently running and holding the lock.

This can happen, for example, if you have a `run_` script that invokes
`chezmoi`, or are running chezmoi in another window.

Under the hood, chezmoi uses [bbolt](https://github.com/etcd-io/bbolt) which
permits multiple simultaneous readers, but only one writer (with no readers).

Commands that take a write lock include `add`, `apply`, `edit`, `forget`,
`import`, `init`, `state`, `unmanage`, and `update`. Commands that take a read
lock include `diff`, `status`, and `verify`.

## chezmoi reports `chezmoi: fork/exec /tmp/XXXXXXXXXX.XX: exec format error` when executing a template script

This error occurs when you have a newline before the `#!` in your script.
Suppress the newline by including a `-` before the closing `}}` on the first
line.

For example, if your template script begins with

```
{{ if eq .chezmoi.os "linux" }}
#!/bin/sh
```

change this to

```
{{ if eq .chezmoi.os "linux" -}}
#!/bin/sh
```

## chezmoi reports `chezmoi: fork/exec /tmp/XXXXXXXXXX.XX: permission denied` when executing a script

This error occurs when your temporary directory is mounted with the `noexec`
option.

As chezmoi scripts can be templates, encrypted, or both, chezmoi needs to write
the final script's contents to a file so that it can be executed by the
operating system. By default, chezmoi will use `$TMPDIR` for this.

You can change the temporary directory into which chezmoi writes and executes
scripts with the `scriptTempDir` configuration variable. For example, to use a
subdirectory of your home directory you can use:

```toml title="~/.config/chezmoi/chezmoi.toml"
scriptTempDir = "~/tmp"
```

## chezmoi reports `chezmoi: mkdir xxxxx: no such file or directory` when trying to manage file or directory

This error occurs when you try to add directory/file to be managed via chezmoi
but the same directory is only listed in `.chezmoiexternal.$FORMAT`.

A workaround can be applied in a such case via manually creating import
directory in chezmoi source directory (typically `~/.local/share/chezmoi`) and
create `.keep` file.

For example, if `.chezmoiexternal.toml` has the configuration:

```toml
[".config/nvim"]
    type = "git-repo"
    url = "https://github.com/NvChad/NvChad.git"
    refreshPeriod = "168h"
    [".config/nvim".pull]
        args = ["--ff-only"]
```

Now `chezmoi add ~/.config/direnv/direnvrc` will raise the error:

```
chezmoi: mkdir /home/<user>/.local/share/chezmoi/dot_config/direnv: no such file or directory
```

But the workaround can be applied:

```console
$ chezmoi cd
$ mkdir -p dot_config/
$ touch dot_config/.keep
```

Now once that done `chezmoi add ~/.config/direnv/direnvrc` should work. For
reference see [this issue](https://github.com/twpayne/chezmoi/issues/2006)

## chezmoi reports `read /dev/stdin: permission denied` or `write /dev/stdout: permission denied` when I redirect standard input or standard output

This error occurs when you [installed chezmoi with
snap](https://snapcraft.io/chezmoi) and is caused by a long-standing [bug in
snap](https://bugs.launchpad.net/ubuntu/+source/snapd/+bug/1849753).

This is not a bug in chezmoi and there is nothing that chezmoi can do about
this. However, there are two workarounds:

Firstly, you can use alternatives to shell redirection. For standard input:

```console
$ chezmoi $COMMAND <$FILENAME       # fails
$ cat $FILENAME | chezmoi $COMMAND  # succeeds
```

For standard output:

```console
$ chezmoi $COMMAND >$FILENAME                  # fails
$ chezmoi $COMMAND -o $FILENAME                # succeeds
$ chezmoi $COMMAND --output=$FILENAME          # succeeds
$ chezmoi $COMMAND | tee $FILENAME >/dev/null  # succeeds
```

Secondly, you can install chezmoi with any of the [many supported install
methods](../../install.md) instead of snap.

## chezmoi reports `fork/exec ...: no such file or directory` when running scripts on Nix or Termux

You are likely using a hardcoded script interpreter in the shebang line of your
scripts, e.g.

```bash
#!/bin/bash
```

`/bin/bash` does not exist on Nix or Termux. You must update the shebang line to point
to the actual bash interpreter. The easiest way to do this is make the script a
template and use the `lookPath` template function, for example:

```
#!{{ lookPath "bash" }}
```

Alternatively, you can use the actual path to `bash` on your system, for example:

=== "Nix"

    ```bash
    #!/usr/bin/env bash
    ```

=== "Termux"

    ```bash
    #!/data/data/com.termux/files/usr/bin/bash
    ```
