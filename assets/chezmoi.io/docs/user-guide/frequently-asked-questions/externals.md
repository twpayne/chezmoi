# Externals

## Why do files in `git-repo` externals appear in `chezmoi unmanaged`?

chezmoi's support for `git-repo` externals is limited to running `git init` and
`git pull` in the directory. This means that the directory is managed by chezmoi
but its contents are not. Consequently, `git-repo` directories are listed by
`chezmoi managed` but their contents are listed in `chezmoi unmanaged`.
