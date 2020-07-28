# chezmoi Quick Start Guide

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

Initialize chezmoi:

    chezmoi init

This will create a new git repository in `~/.local/share/chezmoi` with
permissions `0700` where chezmoi will store the source state.  chezmoi only
modifies files in the working copy. It is your responsibility to commit changes.

Manage an existing file with chezmoi:

    chezmoi add ~/.bashrc

This will copy `~/.bashrc` to `~/.local/share/chezmoi/dot_bashrc`. 
If you want to add a whole folder to chezmoi, you have to add the `-r` argument after `add`.

Edit the source state:

    chezmoi edit ~/.bashrc

This will open `~/.local/share/chezmoi/dot_bashrc` in your `$EDITOR`. Make some
changes and save them.

See what changes chezmoi would make:

    chezmoi diff

Apply the changes:

    chezmoi -v apply

All chezmoi commands accept the `-v` (verbose) flag to print out exactly what
changes they will make to the file system, and the `-n` (dry run) flag to not
make any actual changes. The combination `-n` `-v` is very useful if you want to
see exactly what changes would be made.

Finally, open a shell in the source directory, commit your changes, and return
to where you were:

    chezmoi cd
    git add dot_bashrc
    git commit -m "Add .bashrc"
    exit

## Using chezmoi across multiple machines

Clone the git repo in `~/.local/share/chezmoi` to a hosted Git service, e.g.
[GitHub](https://github.com), [GitLab](https://gitlab.com), or
[BitBucket](https://bitbucket.org). Many people call their dotfiles repo
`dotfiles`. You can then setup chezmoi on a second machine:

    chezmoi init https://github.com/username/dotfiles.git

This will check out the repo and any submodules and optionally create a chezmoi
config file for you. It won't make any changes to your home directory until you
run:

    chezmoi apply

On any machine, you can pull and apply the latest changes from your repo with:

    chezmoi update

## Next steps

For a full list of commands run:

    chezmoi help

chezmoi has much more functionality. Read the [how-to
guide](https://github.com/twpayne/chezmoi/blob/master/docs/HOWTO.md) to explore.
