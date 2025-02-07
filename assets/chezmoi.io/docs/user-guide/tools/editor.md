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

In the specific case of using [VSCode][vscode] or [Codium][codium] as your
editor, you must pass the `--wait` flag, for example, in your shell config:

```bash
export EDITOR="code --wait"
```

Or in chezmoi's configuration file:

```toml title="~/.config/chezmoi/chezmoi.toml"
[edit]
    command = "code"
    args = ["--wait"]
```

!!! warning

    If you use [Helix][helix], you must use Helix 25.01 or later.

## Use chezmoi with VIM

[`github.com/alker0/chezmoi.vim`][alker0] provides syntax highlighting for files
managed by chezmoi, including for templates.

[`github.com/Lilja/vim-chezmoi`][lilja] works with `chezmoi edit` to apply the
edited dotfile on save.

[`github.com/xvzc/chezmoi.nvim`][xvzc] allows you to edit your chezmoi-managed
files and automatically apply.

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

[`github.com/tuh8888/chezmoi.el`][tuh8888] provides convenience functions for
interacting with chezmoi from Emacs, and is available in [MELPA][melpa].

[vscode]: https://code.visualstudio.com/
[codium]: https://vscodium.com/
[alker0]: https://github.com/alker0/chezmoi.vim
[lilja]: https://github.com/Lilja/vim-chezmoi
[xvzc]: https://github.com/xvzc/chezmoi.nvim
[tuh8888]: https://github.com/tuh8888/chezmoi.el
[melpa]: https://melpa.org/#/chezmoi
[helix]: https://helix-editor.com/
