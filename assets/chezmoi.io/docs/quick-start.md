# Quick start

## Concepts

chezmoi stores the desired state of your dotfiles in the directory
`~/.local/share/chezmoi`. When you run `chezmoi apply`, chezmoi calculates the
desired contents and permissions for each dotfile and then makes any changes
necessary so that your dotfiles match that state.

## Start using chezmoi on your current machine

Assuming that you have already [installed chezmoi](/install/), initialize
chezmoi with:

```console
$ chezmoi init
```

This will create a new git repository in `~/.local/share/chezmoi` where chezmoi
will store its source state. By default, chezmoi only modifies files in the
working copy. It is your responsibility to commit and push changes, but chezmoi
can automate this for you if you want.

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
$ git remote add origin git@github.com:username/dotfiles.git
$ git branch -M main
$ git push -u origin main
```

chezmoi can also be used with [GitLab](https://gitlab.com), or
[BitBucket](https://bitbucket.org), [Source Hut](https://sr.ht/), or any other
git hosting service.

Finally, exit the shell in the source directory to return to where you were:

```console
$ exit
```

## Using chezmoi across multiple machines

On a second machine, initialize chezmoi with your dotfiles repo:

```console
$ chezmoi init https://github.com/username/dotfiles.git
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

## Next steps

For a full list of commands run:

```console
$ chezmoi help
```

chezmoi has much more functionality. Good starting points are reading [articles
about chezmoi](/links/articles-podcasts-and-videos/) adding more dotfiles, and
using templates to manage files that vary from machine to machine and retrieve
secrets from your password manager. Read the [user guide](/user-guide/setup/)
to explore.
