# chezmoi

Manage your dotfiles across multiple diverse machines, securely.

With chezmoi, pronounced /ʃeɪ mwa/ (shay-moi), you can install chezmoi and your
dotfiles from your GitHub dotfiles repo on a new, empty machine with a single
command:

```console
$ sh -c "$(curl -fsLS https://chezmoi.io/get)" -- init --apply $GITHUB_USERNAME
```

Updating your dotfiles on any machine is a single command:

```console
$ chezmoi update
```

## How do I start with chezmoi?

[Install chezmoi](/install/) then read the [quick start guide](/quick-start/).
The [user guide](/user-guide/setup/) covers most common tasks. For a full
description of chezmoi, consult the [reference](/reference/).

## Considering using chezmoi?

You can browse other people's dotfiles that use chezmoi [on
GitHub](https://github.com/topics/chezmoi?o=desc&s=updated), [on
GitLab](https://gitlab.com/explore/projects?topic=chezmoi), and [on
Codeberg](https://codeberg.org/explore/repos?sort=recentupdate&q=chezmoi&tab=),
[read articles, listen to podcasts, and watch videos about
chezmoi](/links/articles-podcasts-and-videos/) and see [how chezmoi compares to
other dotfile managers](/comparison-table/).
