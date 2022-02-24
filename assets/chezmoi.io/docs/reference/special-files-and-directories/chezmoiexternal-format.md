# `.chezmoiexternal.<format>`

If a file called `.chezmoiexternal.<format>` exists in the source state, it is
interpreted as a list of external files and archives to be included as if they
were in the source state.

`<format>` must be one of chezmoi's supported configuration file formats, e.g.
`json`, `toml`, or `yaml`.

`.chezmoiexternal.<format>` is interpreted as a template. This allows different
externals to be included on different machines.

Entries are indexed by target name relative to the directory of the
`.chezmoiexternal.<format>` file, and must have a `type` and a `url` field.
`type` can be either `file`, `archive`, or `git-repo`. If the entry's parent
directories do not already exist in the source state then chezmoi will create
them as regular directories.

Entries may have the following fields:

| Variable          | Type     | Default value | Description                                                   |
| ----------------- | -------- | ------------- | ------------------------------------------------------------- |
| `type`            | string   | *none*        | External type (`file`, `archive`, or `git-repo`)              |
| `clone.args`      | []string | *none*        | Extra args to `git clone`                                     |
| `encrypted`       | bool     | `false`       | Whether the external is encrypted                             |
| `exact`           | bool     | `false`       | Add `exact_` attribute to directories in archive              |
| `executable`      | bool     | `false`       | Add `executable_` attribute to file                           |
| `filter.command`  | string   | *none*        | Command to filter contents                                    |
| `filter.args`     | []string | *none*        | Extra args to command to filter contents                      |
| `format`          | string   | *autodetect*  | Format of archive                                             |
| `pull.args`       | []string | *none*        | Extra args to `git pull`                                      |
| `refreshPeriod`   | duration | `0`           | Refresh period                                                |
| `stripComponents` | int      | `0`           | Number of leading directory components to strip from archives |
| `url`             | string   | *none*        | URL                                                           |

The optional boolean `encrypted` field specifies whether the file or archive is
encrypted.

If optional string `filter.command` and array of strings `filter.args` are
specified, the the file or archive is filtered by piping it into the command's
standard input and reading the command's standard output.

If `type` is `file` then the target is a file with the contents of `url`. The
optional boolean field `executable` may be set, in which case the target file
will be executable.

If `type` is `archive` then the target is a directory with the contents of the
archive at `url`. The optional boolean field `exact` may be set, in which case
the directory and all subdirectories will be treated as exact directories, i.e.
`chezmoi apply` will remove entries not present in the archive. The optional
integer field `stripComponents` will remove leading path components from the
members of archive. The optional string field `format` sets the archive format.
The supported archive formats are `tar`, `tar.gz`, `tgz`, `tar.bz2`, `tbz2`,
`xz`, and `zip`. If `format` is not specified then chezmoi will guess the format
using firstly the path of the URL and secondly its contents.

If `type` is `git-repo` then chezmoi will run `git clone <url> <target-name>`
with the optional `clone.args` if the target does not exist. If the target
exists, then chezmoi will run `git pull` with the optional `pull.args` to
update the target.

For `file` and `archive` externals, chezmoi will cache downloaded URLs. The
optional duration `refreshPeriod` field specifies how often chezmoi will
re-download the URL. The default is zero meaning that chezmoi will never
re-download unless forced. To force chezmoi to re-download URLs, pass the
`-R`/`--refresh-externals` flag. Suitable refresh periods include one day
(`24h`), one week (`168h`), or four weeks (`672h`).

!!! example

    ```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
    [".vim/autoload/plug.vim"]
        type = "file"
        url = "https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim"
        refreshPeriod = "168h"
    [".oh-my-zsh"]
        type = "archive"
        url = "https://github.com/ohmyzsh/ohmyzsh/archive/master.tar.gz"
        exact = true
        stripComponents = 1
        refreshPeriod = "168h"
    [".oh-my-zsh/custom/plugins/zsh-syntax-highlighting"]
        type = "archive"
        url = "https://github.com/zsh-users/zsh-syntax-highlighting/archive/master.tar.gz"
        exact = true
        stripComponents = 1
        refreshPeriod = "168h"
    [".oh-my-zsh/custom/themes/powerlevel10k"]
        type = "archive"
        url = "https://github.com/romkatv/powerlevel10k/archive/v1.15.0.tar.gz"
        exact = true
        stripComponents = 1
    ```
