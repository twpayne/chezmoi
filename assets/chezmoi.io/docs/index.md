# chezmoi

Manage your dotfiles across multiple diverse machines, securely.

chezmoi helps you manage your personal configuration files (dotfiles, like
`~/.gitconfig`) across multiple machines.

chezmoi provides many features beyond symlinking or using a bare git repo
including: templates (to handle small differences between machines), password
manager support (to store your secrets securely), importing files from archives
(great for shell and editor plugins), full file encryption (using gpg or age),
and running scripts (to handle everything else).

With chezmoi, pronounced /ʃeɪ mwa/ (shay-moi), you can install chezmoi and your
dotfiles from your GitHub dotfiles repo on a new, empty machine with a single
command:

```console
$ sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply $GITHUB_USERNAME
```

As well as the `curl | sh` installation, you can [install chezmoi with your
favorite package manager](/install/).

Updating your dotfiles on any machine is a single command:

```console
$ chezmoi update
```

## How to start with chezmoi?

[Install chezmoi](/install/) then read the [quick start guide](/quick-start/).
The [user guide](/user-guide/setup/) covers most common tasks. For a full
description, consult the [reference](/reference/).

## Not sure if you should use chezmoi?

[Read articles, listen to podcasts, and watch videos about
chezmoi](/links/articles-podcasts-and-videos/) and [compare chezmoi to other
dotfile managers](/comparison-table/).

See how other people use chezmoi by [browsing chezmoi dotfile repos on
GitHub](https://github.com/topics/chezmoi?o=desc&s=updated),
[GitLab](https://gitlab.com/explore/projects?topic=chezmoi), and
[Codeberg](https://codeberg.org/explore/repos?sort=recentupdate&q=chezmoi&tab=).

## Like chezmoi and want to say thanks?

Please [give chezmoi a star on
GitHub](https://github.com/twpayne/chezmoi/stargazers).

[Share chezmoi](/links/articles-podcasts-and-videos/) and, if you're happy to
share your public dotfiles repo, then [tag your repo with
`chezmoi`](https://github.com/topics/chezmoi?o=desc&s=updated).

[Contributions are very welcome](/developer/contributing-changes/) and every
[bug report, support request, and feature
request](https://github.com/twpayne/chezmoi/issues/new/choose) helps make
chezmoi better. Thank you :)

If you would like to make a financial contribution, please instead make a
donation to a charity or cause of your choice.
