# Include dotfiles from elsewhere

## Include a subdirectory from another repository, like Oh My Zsh

To include a subdirectory from another repository, e.g. [Oh My
Zsh](https://github.com/ohmyzsh/ohmyzsh), you cannot use git submodules because
chezmoi uses its own format for the source state and Oh My Zsh is not
distributed in this format. Instead, you can use the
`.chezmoiexternal.<format>` to tell chezmoi to import dotfiles from an external
source.

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

## Include a single file from another repository

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

Note that this only updates the source state. You will need to run

```console
$ chezmoi apply
```

to update your destination directory.
