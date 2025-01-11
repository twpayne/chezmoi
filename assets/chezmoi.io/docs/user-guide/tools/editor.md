# Editor

## Use your preferred editor with `chezmoi edit` and `chezmoi edit-config`

By default, chezmoi will use your preferred editor as defined by the `$VISUAL`
or `$EDITOR` environment variables, falling back to a default editor depending
on your operating system (`vi` on UNIX-like operating systems, `notepad.exe` on
Windows).

You can configure chezmoi to use your preferred editor by either setting the
`$EDITOR` environment variable or setting the `edit.command` variable in your
configuration file.

The editor command must only return when you have finished editing the files.
chezmoi will emit a warning if your editor command returns too quickly.

In the specific case of using [VSCode](https://code.visualstudio.com/) or
[Codium](https://vscodium.com/) as your editor, you must pass the `--wait`
flag, for example, in your shell config:

```bash
export EDITOR="code --wait"
```

Or in chezmoi's configuration file:

```toml title="~/.config/chezmoi/chezmoi.toml"
[edit]
    command = "code"
    args = ["--wait"]
```

## Use chezmoi with VIM

[`github.com/alker0/chezmoi.vim`](https://github.com/alker0/chezmoi.vim)
provides syntax highlighting for files managed by chezmoi, including for
templates.

[`github.com/Lilja/vim-chezmoi`](https://github.com/Lilja/vim-chezmoi) works
with `chezmoi edit` to apply the edited dotfile on save.

[`github.com/xvzc/chezmoi.nvim`](https://github.com/xvzc/chezmoi.nvim) allows
you to edit your chezmoi-managed files and automatically apply.

Alternatively, you can use an `autocmd` to run `chezmoi apply` whenever you save
a dotfile, but you must disable `chezmoi edit`'s hardlinking:

```toml title="~/.config/chezmoi/chezmoi.toml"
[edit]
    hardlink = false
```

```vim title="~/.vimrc"
autocmd BufWritePost ~/.local/share/chezmoi/* ! chezmoi apply --source-path "%"
```

## Use chezmoi with emacs

[`github.com/tuh8888/chezmoi.el`](https://github.com/tuh8888/chezmoi.el)
provides convenience functions for interacting with chezmoi from emacs, and is
available in [MELPA](https://melpa.org/#/chezmoi).
