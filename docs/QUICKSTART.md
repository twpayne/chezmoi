# chezmoi quick start guide

<!--- toc --->
* [Concepts](#concepts)
* [Start using chezmoi on your current machine](#start-using-chezmoi-on-your-current-machine)
* [Using chezmoi across multiple machines](#using-chezmoi-across-multiple-machines)
* [Next steps](#next-steps)

## Concepts

chezmoi stores the desired state of your dotfiles in the directory
`~/.local/share/chezmoi`. When you run `chezmoi apply`, chezmoi calculates the
desired contents and permissions for each dotfile and then makes any changes
necessary so that your dotfiles match that state.

## Start using chezmoi on your current machine

Assuming that you have already [installed
chezmoi](https://github.com/twpayne/chezmoi/blob/master/docs/INSTALL.md),
initialize chezmoi with:

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
config file for you. It won't make any changes to your home directory until you
run:

```console
$ chezmoi apply
```

If your dotfiles repo is `https://github.com/username/dotfiles.git` then the
above two commands can be combined into just:

```console
$ chezmoi init --apply username
```

On any machine, you can pull and apply the latest changes from your repo with:

```console
$ chezmoi update
```

## Next steps

For a full list of commands run:

```console
$ chezmoi help
```

chezmoi has much more functionality. Good starting points are adding more
dotfiles, and using templates to manage files that vary from machine to machine
and retrieve secrets from your password manager. Read the [how-to
guide](https://github.com/twpayne/chezmoi/blob/master/docs/HOWTO.md) to explore.
