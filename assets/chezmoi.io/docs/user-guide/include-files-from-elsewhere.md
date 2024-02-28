# Include dotfiles from elsewhere

The sections below contain examples of how to use `.chezmoiexternal.toml` to
include files from external sources. For more details, check the [reference
manual](../reference/special-files-and-directories/chezmoiexternal-format.md) .

## Include a subdirectory from a URL

To include a subdirectory from another repository, e.g. [Oh My
Zsh](https://github.com/ohmyzsh/ohmyzsh), you cannot use git submodules because
chezmoi uses its own format for the source state and Oh My Zsh is not
distributed in this format. Instead, you can use the `.chezmoiexternal.$FORMAT`
to tell chezmoi to import dotfiles from an external source.

For example, to import Oh My Zsh, the [zsh-syntax-highlighting
plugin](https://github.com/zsh-users/zsh-syntax-highlighting), and
[powerlevel10k](https://github.com/romkatv/powerlevel10k), put the following in
`~/.local/share/chezmoi/.chezmoiexternal.toml`:

```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
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

To apply the changes, run:

```console
$ chezmoi apply
```

chezmoi will download the archives and unpack them as if they were part of the
source state. chezmoi caches downloaded archives locally to avoid
re-downloading them every time you run a chezmoi command, and will only
re-download them at most every `refreshPeriod` (default never).

In the above example `refreshPeriod` is set to `168h` (one week) for
`.oh-my-zsh` and `.oh-my-zsh/custom/plugins/zsh-syntax-highlighting` because
the URL point to tarballs of the `master` branch, which changes over time. No
refresh period is set for `.oh-my-zsh/custom/themes/powerlevel10k` because the
URL points to the a tarball of a tagged version, which does not change over
time. To bump the version of powerlevel10k, change the version in the URL.

To force a refresh the downloaded archives, use the `--refresh-externals` flag
to `chezmoi apply`:

```console
$ chezmoi --refresh-externals apply
```

`--refresh-externals` can be shortened to `-R`:

```console
$ chezmoi -R apply
```

When using Oh My Zsh, make sure you disable auto-updates by setting
`DISABLE_AUTO_UPDATE="true"` in `~/.zshrc`. Auto updates will cause the
`~/.oh-my-zsh` directory to drift out of sync with chezmoi's source state. To
update Oh My Zsh and its plugins, refresh the downloaded archives.

!!! note

    If your external dependency target directory can contain cache files that are
    added during normal use, chezmoi will report that files have changed on `chezmoi
    apply`. To avoid this, add the cache directory to your
    [`.chezmoiignore`](../reference/special-files-and-directories/chezmoiignore.md)
    file.

    For example, Oh My Zsh may cache completions in `.oh-my-zsh/cache/completions/`,
    which should be added to your `.chezmoiignore` file.

## Include a subdirectory with selected files from a URL

Use `include` pattern filters to include only selected files from an archive
URL.

For example, to import just the required source files of the
[zsh-syntax-highlighting
plugin](https://github.com/zsh-users/zsh-syntax-highlighting) in the example
above, add in `include` filter to the `zsh-syntax-highlighting` section as shown
below:

```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
[".oh-my-zsh/custom/plugins/zsh-syntax-highlighting"]
    type = "archive"
    url = "https://github.com/zsh-users/zsh-syntax-highlighting/archive/master.tar.gz"
    exact = true
    stripComponents = 1
    refreshPeriod = "168h"
    include = ["*/*.zsh", "*/.version", "*/.revision-hash", "*/highlighters/**"]
```

## Include a single file from a URL

Including single files uses the same mechanism as including a subdirectory
above, except with the external type `file` instead of `archive`. For example,
to include
[`plug.vim`](https://github.com/junegunn/vim-plug/blob/master/plug.vim) from
[`github.com/junegunn/vim-plug`](https://github.com/junegunn/vim-plug) in
`~/.vim/autoload/plug.vim` put the following in
`~/.local/share/chezmoi/.chezmoiexternal.toml`:

```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
[".vim/autoload/plug.vim"]
    type = "file"
    url = "https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim"
    refreshPeriod = "168h"
```

## Extract a single file from an archive

You can extract a single file from an archive using the `archive-file` type in
`.chezmoiexternal.$FORMAT`, for example:

```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
{{ $ageVersion := "1.1.1" -}}
[".local/bin/age"]
    type = "archive-file"
    url = "https://github.com/FiloSottile/age/releases/download/v{{ $ageVersion }}/age-v{{ $ageVersion }}-{{ .chezmoi.os }}-{{ .chezmoi.arch }}.tar.gz"
    path = "age/age"
```

This will extract the single archive member `age/age` from the given URL (which
is computed for the current OS and architecture) to the target
`./local/bin/age`.

## Import archives

It is occasionally useful to import entire archives of configuration into your
source state. The `import` command does this. For example, to import the latest
version [`github.com/ohmyzsh/ohmyzsh`](https://github.com/ohmyzsh/ohmyzsh) to
`~/.oh-my-zsh` run:

```console
$ curl -s -L -o ${TMPDIR}/oh-my-zsh-master.tar.gz https://github.com/ohmyzsh/ohmyzsh/archive/master.tar.gz
$ mkdir -p $(chezmoi source-path)/dot_oh-my-zsh
$ chezmoi import --strip-components 1 --destination ~/.oh-my-zsh ${TMPDIR}/oh-my-zsh-master.tar.gz
```

!!! note

    This only updates the source state. You will need to run:

    ```console
    $ chezmoi apply
    ```

    to update your destination directory.

## Handle tar archives in an unsupported compression format

chezmoi natively understands tar archives. tar archives can be uncompressed or
compressed in the bzip2, gzip, xz, or zstd formats.

If you have a tar archive in an unsupported compression format then you can use
a filter to decompress it. For example, before chezmoi natively supported the
zstd compression format, you could handle `.tar.zst` external archives with, for
example:

```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
[".Software/anki/2.1.54-qt6"]
    type = "archive"
    url = "https://github.com/ankitects/anki/releases/download/2.1.54/anki-2.1.54-linux-qt6.tar.zst"
    filter.command = "zstd"
    filter.args = ["-d"]
    format = "tar"
```

Here `filter.command` and `filter.args` together tell chezmoi to filter the
downloaded data through `zstd -d`. The `format = "tar"` line tells chezmoi that
output of the filter is an uncompressed tar archive.

## Include a subdirectory from a git repository

You can configure chezmoi to keep a git repository up to date in a subdirectory
by using the external type `git-repo`, for example:

```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
[".vim/pack/alker0/chezmoi.vim"]
    type = "git-repo"
    url = "https://github.com/alker0/chezmoi.vim.git"
    refreshPeriod = "168h"
```

If the directory does not exist then chezmoi will run `git clone` to clone it.
If the directory does exist then chezmoi will run `git pull` to pull the latest
changes, but not more often than every `refreshPeriod`. In the above example
the `refreshPeriod` is `168h` which is one week. The default `refreshPeriod` is
zero, which disables refreshes. You can force a refresh (i.e. force a `git
pull`) by passing the `--refresh-externals`/`-R` flag to `chezmoi apply`.

!!! warning

    chezmoi's support for `git-repo` externals is limited to running `git
    clone` and/or `git pull` in a directory. You must have a `git` binary
    in your `$PATH`.

    Using a `git-repo` external delegates management of the
    directory to git. chezmoi cannot manage any other files in that directory.

    The contents of `git-repo` externals will not be manifested in commands
    like `chezmoi diff` or `chezmoi dump`, and will be listed by `chezmoi
    unmanaged`.

!!! hint

    If you need to manage extra files in a `git-repo` external, use an
    `archive` external instead with the URL pointing to an archive of the git
    repo's `master` or `main` branch.

You can customize the arguments to `git clone` and `git pull` by setting the
`$DIR.clone.args` and `$DIR.pull.args` variables in `.chezmoiexternal.$FORMAT`,
for example:

```toml title="~/.local/share/chezmoi/.chezmoiexternal.toml"
[".vim/pack/alker0/chezmoi.vim"]
    type = "git-repo"
    url = "https://github.com/alker0/chezmoi.vim.git"
    refreshPeriod = "168h"
    [".vim/pack/alker0/chezmoi.vim".pull]
        args = ["--ff-only"]
```

## Use git submodules in your source directory

!!! important

    If you use git submodules, then you should set the `external_` attribute on
    the subdirectory containing the submodule.

You can include git repos from elsewhere as git submodules in your source
directory. `chezmoi init` and `chezmoi update` are aware of git submodules and
will run git with the `--recurse-submodules` flag by default.

chezmoi assumes that all files and directories in its source state are in
chezmoi's format, i.e. their filenames include attributes like `private_` and
`run_`.  Most git submodules are not in chezmoi's format and so files like
`run_test.sh` will be interpreted by chezmoi as a `run_` script. To avoid
this problem, set the `external_` attribute on all subdirectories that contain
submodules.

You can stop chezmoi from handling git submodules by passing the
`--recurse-submodules=false` flag or setting the `update.recurseSubmodules`
configuration variable to `false`.
