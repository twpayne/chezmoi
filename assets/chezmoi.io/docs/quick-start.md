# Quick start

## Concepts

Roughly speaking, chezmoi stores the desired state of your dotfiles in the
directory `~/.local/share/chezmoi`. When you run `chezmoi apply`, chezmoi
calculates the desired contents for each of your dotfiles and then makes the
minimum changes required to make your dotfiles match your desired state.
chezmoi's concepts are [described more accurately in the reference
manual](reference/concepts.md).

## Start using chezmoi on your current machine

Assuming that you have already [installed chezmoi](install.md), initialize
chezmoi with:

```console
$ chezmoi init
```

This will create a new git local repository in `~/.local/share/chezmoi` where
chezmoi will store its source state. By default, chezmoi only modifies files in
the working copy.

Manage your first file with chezmoi:

```console
$ chezmoi add ~/.bashrc
```

This will copy `~/.bashrc` to `~/.local/share/chezmoi/dot_bashrc`.

Edit the source state:

```console
$ chezmoi edit ~/.bashrc
```

This will open `~/.local/share/chezmoi/dot_bashrc` in your `$EDITOR`. Make some
changes and save the file.

!!! hint

    You don't have to use `chezmoi edit` to edit your dotfiles. See [this FAQ
    entry](user-guide/frequently-asked-questions/usage.md#how-do-i-edit-my-dotfiles-with-chezmoi)
    for more details.

See what changes chezmoi would make:

```console
$ chezmoi diff
```

Apply the changes:

```console
$ chezmoi -v apply
```

All chezmoi commands accept the `-v` (verbose) flag to print out exactly what
changes they will make to the file system, and the `-n` (dry run) flag to not
make any actual changes. The combination `-n` `-v` is very useful if you want to
see exactly what changes would be made.

Next, open a shell in the source directory, to commit your changes:

```console
$ chezmoi cd
$ git add .
$ git commit -m "Initial commit"
```

[Create a new repository on GitHub](https://github.com/new) called `dotfiles`
and then push your repo:

```console
$ git remote add origin https://github.com/$GITHUB_USERNAME/dotfiles.git
$ git branch -M main
$ git push -u origin main
```

!!! hint

    chezmoi can be configured to automatically add, commit, and push changes to
    your repo.

chezmoi can also be used with [GitLab](https://gitlab.com), or
[BitBucket](https://bitbucket.org), [Source Hut](https://sr.ht/), or any other
git hosting service.

Finally, exit the shell in the source directory to return to where you were:

```console
$ exit
```

These commands are summarized in this sequence diagram:

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    H->>L: chezmoi init
    H->>W: chezmoi add &lt;file&gt;
    W->>W: chezmoi edit &lt;file&gt;
    W-->>H: chezmoi diff
    W->>H: chezmoi apply
    H-->>W: chezmoi cd
    W->>L: git add
    W->>L: git commit
    L->>R: git push
    W-->>H: exit
```

## Using chezmoi across multiple machines

On a second machine, initialize chezmoi with your dotfiles repo:

```console
$ chezmoi init https://github.com/$GITHUB_USERNAME/dotfiles.git
```

!!! hint

    Private GitHub repos require other
    [authentication methods](https://docs.github.com/en/get-started/getting-started-with-git/about-remote-repositories#cloning-with-https-urls):

    ```console
    $ chezmoi init git@github.com:$GITHUB_USERNAME/dotfiles.git
    ```

This will check out the repo and any submodules and optionally create a chezmoi
config file for you.

Check what changes that chezmoi will make to your home directory by running:

```console
$ chezmoi diff
```

If you are happy with the changes that chezmoi will make then run:

```console
$ chezmoi apply -v
```

If you are not happy with the changes to a file then either edit it with:

```console
$ chezmoi edit $FILE
```

Or, invoke a merge tool (by default `vimdiff`) to merge changes between the
current contents of the file, the file in your working copy, and the computed
contents of the file:

```console
$ chezmoi merge $FILE
```

On any machine, you can pull and apply the latest changes from your repo with:

```console
$ chezmoi update -v
```

These commands are summarized in this sequence diagram:

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    R->>W: chezmoi init &lt;repo&gt;
    W-->>H: chezmoi diff
    W->>H: chezmoi apply
    W->>W: chezmoi edit &lt;file&gt;
    W->>W: chezmoi merge &lt;file&gt;
    R->>H: chezmoi update
```

## Set up a new machine with a single command

You can install your dotfiles on new machine with a single command:

```console
$ chezmoi init --apply https://github.com/$GITHUB_USERNAME/dotfiles.git
```

If you use GitHub and your dotfiles repo is called `dotfiles` then this can be
shortened to:

```console
$ chezmoi init --apply $GITHUB_USERNAME
```

!!! hint

    Private GitHub repos require other
    [authentication methods](https://docs.github.com/en/get-started/getting-started-with-git/about-remote-repositories#cloning-with-https-urls):

    ```console
    chezmoi init --apply git@github.com:$GITHUB_USERNAME/dotfiles.git
    ```

This command is summarized in this sequence diagram:

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    R->>H: chezmoi init --apply &lt;repo&gt;
```

## Next steps

For a full list of commands run:

```console
$ chezmoi help
```

chezmoi has much more functionality. Good starting points are reading [what
other people say about chezmoi](links/articles.md), adding more dotfiles, and
using templates to manage files that vary from machine to machine and retrieve
secrets from your password manager. Read the [user guide](user-guide/setup.md)
to explore and see [how people use chezmoi](links/dotfile-repos.md) for
inspiration.
