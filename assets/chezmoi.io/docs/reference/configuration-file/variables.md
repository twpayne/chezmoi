# Variables

The following configuration variables are available:

| Section        | Variable              | Type     | Default value            | Description                                            |
| -------------- | --------------------- | -------- | ------------------------ | ------------------------------------------------------ |
| Top level      | `color`               | string   | `auto`                   | Colorize output                                        |
|                | `data`                | any      | *none*                   | Template data                                          |
|                | `destDir`             | string   | `~`                      | Destination directory                                  |
|                | `encryption`          | string   | *none*                   | Encryption tool, either `age` or `gpg`                 |
|                | `format`              | string   | `json`                   | Format for data output, either `json` or `yaml`        |
|                | `mode`                | string   | `file`                   | Mode in target dir, either `file` or `symlink`         |
|                | `sourceDir`           | string   | `~/.local/share/chezmoi` | Source directory                                       |
|                | `pager`               | string   | `$PAGER`                 | Default pager                                          |
|                | `umask`               | int      | *from system*            | Umask                                                  |
|                | `useBuiltinAge`       | string   | `auto`                   | Use builtin age if `age` command is not found in $PATH |
|                | `useBuiltinGit`       | string   | `auto`                   | Use builtin git if `git` command is not found in $PATH |
|                | `verbose`             | bool     | `false`                  | Make output more verbose                               |
|                | `workingTree`         | string   | *source directory*       | git working tree directory                             |
| `add`          | `templateSymlinks`    | bool     | `false`                  | Template symlinks to source and home dirs              |
| `age`          | `args`                | []string | *none*                   | Extra args to age CLI command                          |
|                | `command`             | string   | `age`                    | age CLI command                                        |
|                | `identity`            | string   | *none*                   | age identity file                                      |
|                | `identities`          | []string | *none*                   | age identity files                                     |
|                | `passphrase`          | bool     | `false`                  | Use age passphrase instead of identity                 |
|                | `recipient`           | string   | *none*                   | age recipient                                          |
|                | `recipients`          | []string | *none*                   | age recipients                                         |
|                | `recipientsFile`      | []string | *none*                   | age recipients file                                    |
|                | `recipientsFiles`     | []string | *none*                   | age recipients files                                   |
|                | `suffix`              | string   | `.age`                   | Suffix appended to age-encrypted files                 |
|                | `symmetric`           | bool     | `false`                  | Use age symmetric encryption                           |
| `bitwarden`    | `command`             | string   | `bw`                     | Bitwarden CLI command                                  |
| `cd`           | `args`                | []string | *none*                   | Extra args to shell in `cd` command                    |
|                | `command`             | string   | *none*                   | Shell to run in `cd` command                           |
| `completion`   | `custom`              | bool     | `false`                  | Enable custom shell completions                        |
| `diff`         | `args`                | []string | *see `diff` below*       | Extra args to external diff command                    |
|                | `command`             | string   | *none*                   | External diff command                                  |
|                | `exclude`             | []string | *none*                   | Entry types to exclude from diffs                      |
|                | `pager`               | string   | *none*                   | Diff-specific pager                                    |
| `edit`         | `args`                | []string | *none*                   | Extra args to edit command                             |
|                | `command`             | string   | `$EDITOR` / `$VISUAL`    | Edit command                                           |
|                | `hardlink`            | bool     | `true`                   | Invoke editor with a hardlink to the source file       |
|                | `minDuration`         | duration | `1s`                     | Minimum duration for edit command                      |
| `secret`       | `command`             | string   | *none*                   | Generic secret command                                 |
| `git`          | `autoAdd `            | bool     | `false`                  | Add changes to the source state after any change       |
|                | `autoCommit`          | bool     | `false`                  | Commit changes to the source state after any change    |
|                | `autoPush`            | bool     | `false`                  | Push changes to the source state after any change      |
|                | `command`             | string   | `git`                    | Source version control system                          |
| `gopass`       | `command`             | string   | `gopass`                 | gopass CLI command                                     |
| `gpg`          | `args`                | []string | *none*                   | Extra args to GPG CLI command                          |
|                | `command`             | string   | `gpg`                    | GPG CLI command                                        |
|                | `recipient`           | string   | *none*                   | GPG recipient                                          |
|                | `suffix`              | string   | `.asc`                   | Suffix appended to GPG-encrypted files                 |
|                | `symmetric`           | bool     | `false`                  | Use symmetric GPG encryption                           |
| `interpreters` | *extension*`.args`    | []string | *none*                   | See section on "Scripts on Windows"                    |
|                | *extension*`.command` | string   | *special*                | See section on "Scripts on Windows"                    |
| `keepassxc`    | `args`                | []string | *none*                   | Extra args to KeePassXC CLI command                    |
|                | `command`             | string   | `keepassxc-cli`          | KeePassXC CLI command                                  |
|                | `database`            | string   | *none*                   | KeePassXC database                                     |
| `lastpass`     | `command`             | string   | `lpass`                  | Lastpass CLI command                                   |
| `merge`        | `args`                | []string | *see `merge` below*      | Args to 3-way merge command                            |
|                | `command`             | string   | `vimdiff`                | 3-way merge command                                    |
| `onepassword`  | `cache`               | bool     | `true`                   | Enable optional caching provided by `op`               |
|                | `command`             | string   | `op`                     | 1Password CLI command                                  |
|                | `prompt`              | bool     | `true`                   | Prompt for sign-in when no valid session is available  |
| `pass`         | `command`             | string   | `pass`                   | Pass CLI command                                       |
| `pinentry`     | `args`                | []string | *none*                   | Extra args to the pinentry command                     |
|                | `command`             | string   | *none*                   | pinentry command                                       |
|                | `options`             | []string | *see `pinentry` below*   | Extra options for pinentry                             |
| `status`       | `exclude`             | []string | *none*                   | Entry types to exclude from status                     |
| `template`     | `options`             | []string | `["missingkey=error"]`   | Template options                                       |
| `vault`        | `command`             | string   | `vault`                  | Vault CLI command                                      |
