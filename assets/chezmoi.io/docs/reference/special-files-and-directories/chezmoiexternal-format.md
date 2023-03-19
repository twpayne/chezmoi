# `.chezmoiexternal.$FORMAT{,.tmpl}`

If a file called `.chezmoiexternal.$FORMAT` (with an optional `.tmpl` extension)
exists in the source state (either `~/.local/share/chezmoi` or directory defined
inside `.chezmoiroot`), it is interpreted as a list of external files and
archives to be included as if they were in the source state.

`$FORMAT` must be one of chezmoi's supported configuration file formats, e.g.
`json`, `jsonc`, `toml`, or `yaml`.

`.chezmoiexternal.$FORMAT` is interpreted as a template. This allows different
externals to be included on different machines.

Entries are indexed by target name relative to the directory of the
`.chezmoiexternal.$FORMAT` file, and must have a `type` and a `url` field.
`type` can be either `file`, `archive`, or `git-repo`. If the entry's parent
directories do not already exist in the source state then chezmoi will create
them as regular directories.

Entries may have the following fields:

| Variable          | Type     | Default value | Description                                                   |
| ----------------- | -------- | ------------- | ------------------------------------------------------------- |
| `type`            | string   | *none*        | External type (`file`, `archive`, or `git-repo`)              |
| `encrypted`       | bool     | `false`       | Whether the external is encrypted                             |
| `exact`           | bool     | `false`       | Add `exact_` attribute to directories in archive              |
| `exclude`         | []string | *none*        | Patterns to exclude from archive                              |
| `executable`      | bool     | `false`       | Add `executable_` attribute to file                           |
| `format`          | string   | *autodetect*  | Format of archive                                             |
| `include`         | []string | *none*        | Patterns to include from archive                              |
| `refreshPeriod`   | duration | `0`           | Refresh period                                                |
| `stripComponents` | int      | `0`           | Number of leading directory components to strip from archives |
| `url`             | string   | *none*        | URL                                                           |
| `checksum.sha256` | string   | *none*        | Expected SHA256 checksum of data                              |
| `checksum.sha384` | string   | *none*        | Expected SHA384 checksum of data                              |
| `checksum.sha512` | string   | *none*        | Expected SHA512 checksum of data                              |
| `checksum.size`   | int      | *none*        | Expected size of data                                         |
| `clone.args`      | []string | *none*        | Extra args to `git clone`                                     |
| `filter.command`  | string   | *none*        | Command to filter contents                                    |
| `filter.args`     | []string | *none*        | Extra args to command to filter contents                      |
| `pull.args`       | []string | *none*        | Extra args to `git pull`                                      |

If any of the optional `checksum.sha256`, `checksum.sha384`, or
`checksum.sha512` fields are set, chezmoi will verify that the downloaded data
has the given checksum.

The optional boolean `encrypted` field specifies whether the file or archive is
encrypted.

If optional string `filter.command` and array of strings `filter.args` are
specified, the file or archive is filtered by piping it into the command's
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
`xz`, `.tar.zst`, and `zip`. If `format` is not specified then chezmoi will
guess the format using firstly the path of the URL and secondly its contents.

The optional `include` and `exclude` fields are lists of patterns specify which
archive members to include or exclude respectively. Patterns match paths in the
archive, not the target state. chezmoi uses the following algorithm to
determine whether an archive member is included:

1. If the archive member name matches any `exclude` pattern, then the archive
   member is excluded. In addition, if the archive member is a directory, then
   all contained files and sub-directories will be excluded, too (recursively).
2. Otherwise, if the archive member name matches any `include` pattern, then
   the archive member is included.
3. Otherwise, if only `include` patterns were specified then the archive member
   is excluded.
4. Otherwise, if only `exclude` patterns were specified then the archive member
   is included.
5. Otherwise, the archive member is included.o

Excluded archive members do not generate source state entries, and, if they are
directories, all of their children are also excluded.

If `type` is `git-repo` then chezmoi will run `git clone $URL $TARGET_NAME`
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
    ["www/adminer/plugins"]
        type = "archive"
        url = "https://api.github.com/repos/vrana/adminer/tarball"
        refreshPeriod = "744h"
        stripComponents = 2
        include = ["*/plugins/**"]
    ```

    Some more examples can be found in the
    [user guide](/user-guide/include-files-from-elsewhere/).
