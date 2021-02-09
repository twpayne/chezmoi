## Changes in v2, already done

General:
- `--recursive` is default for some commands, notably `chezmoi add`
- only diff format is git
- remove hg support
- remove source command (use git instead)
- `--include` option to many commands
- errors output to stderr, not stdout
- `--force` now global
- `--output` now global
- diff includes scripts
- archive includes scripts
- `encrypt` -> `encrypted` in chattr
- `--format` now global, don't use toml for dump
- `y`, `yes`, `on`, `n`, `no`, `off` recognized as bools
- order for `merge` is now dest, target, source
- No more `--prompt` to `chezmoi edit`
- `--keep-going` global
- `chezmoi init` guesses your repo URL if you use github.com and dotfiles
- `edit.command` and `edit.args` settable in config file, overrides `$EDITOR` / `$VISUAL`
- state data has changed, `run_once_` scripts will be run again
- `init` gets `--depth` and `--purge`
- `run_once_` scripts with same content but different names will only be run once
- global `--use-builtin-git`
- added `.chezmoi.version` template var
- added `gitHubKeys` template func uses `CHEZMOI_GITHUB_ACCESS_TOKEN`, `GITHUB_ACCESS_TOKEN`, and `GITHUB_TOKEN` first non-empty
- template data on a best-effort basis, errors ignored
- `chezmoi status`
- `chezmoi apply` no longer overwrites by default
- `chezmoi init --one-shot`
- new type `--create`
- `chezmoi archive --format=zip`
- `before_` and `after_` script attributes change script order, scripts now run during
- new `fqdnHostname` template var (UNIX only for now)
- age encryption support

{{ range (gitHubKeys "twpayne") -}}
{{ .Key }}
{{ end -}}

Config file:
- rename `sourceVCS` to `git`
- use `gpg.recipient` instead of `gpgRecipient`
- rename `genericSecret` to `secret`
- rename `homedir` to `homeDir`
- add `encryption` (currently `age` or `gpg`)
- apply `--ignore-encrypted`
- apply `--source-paths`
