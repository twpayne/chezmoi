# Usage

## How do I edit my dotfiles with chezmoi?

There are four popular approaches:

1. Use `chezmoi edit $FILE`. This will open the source file for `$FILE` in your
   editor, including . For extra ease, use `chezmoi edit --apply $FILE` to apply
   the changes when you quit your editor.
2. Use `chezmoi cd` and edit the files in the source directory directly. Run
   `chezmoi diff` to see what changes would be made, and `chezmoi apply` to make
   the changes.
3. If your editor supports opening directories, run `chezmoi edit` with no
   arguments to open the source directory.
4. Edit the file in your home directory, and then either re-add it by running
   `chezmoi add $FILE` or `chezmoi re-add`. Note that `re-add` doesn't work with
   templates.
5. Edit the file in your home directory, and then merge your changes with source
   state by running `chezmoi merge $FILE`.

## What are the consequences of "bare" modifications to the target files? If my `.zshrc` is managed by chezmoi and I edit `~/.zshrc` without using `chezmoi edit`, what happens?

Until you run `chezmoi apply` your modified `~/.zshrc` will remain in place.
When you run `chezmoi apply` chezmoi will detect that `~/.zshrc` has changed
since chezmoi last wrote it and prompt you what to do. You can resolve
differences with a merge tool by running `chezmoi merge ~/.zshrc`.

## How can I tell what dotfiles in my home directory aren't managed by chezmoi? Is there an easy way to have chezmoi manage a subset of them?

`chezmoi unmanaged` will list everything not managed by chezmoi. You can add
entire directories with `chezmoi add`.

## How can I tell what dotfiles in my home directory are currently managed by chezmoi?

`chezmoi managed` will list everything managed by chezmoi.

## If there's a mechanism in place for the above, is there also a way to tell chezmoi to ignore specific files or groups of files (e.g. by directory name or by glob)?

By default, chezmoi ignores everything that you haven't explicitly added. If you
have files in your source directory that you don't want added to your
destination directory when you run `chezmoi apply` add their names to a file
called `.chezmoiignore` in the source state.

Patterns are supported, and you can change what's ignored from machine to
machine. The full usage and syntax is described in the [reference
manual](/reference/special-files-and-directories/chezmoiignore/).

## If the target already exists, but is "behind" the source, can chezmoi be configured to preserve the target version before replacing it with one derived from the source?

Yes. Run `chezmoi add` will update the source state with the target. To see
diffs of what would change, without actually changing anything, use `chezmoi
diff`.

## Once I've made a change to the source directory, how do I commit it?

You have several options:

* `chezmoi cd` opens a shell in the source directory, where you can run your
  usual version control commands, like `git add` and `git commit`.

* `chezmoi git` runs `git` in the source
  directory and pass extra arguments to the command. If you're passing any
  flags, you'll need to use `--` to prevent chezmoi from consuming them, for
  example `chezmoi git -- commit -m "Update dotfiles"`.

* You can configure chezmoi to automatically commit and push changes to your
  source state, as [described in the how-to
  guide](/user-guide/daily-operations/#automatically-commit-and-push-changes-to-your-repo).

## I've made changes to both the destination state and the source state that I want to keep. How can I keep them both?

`chezmoi merge` will open a merge tool to resolve differences between the source
state, target state, and destination state. Copy the changes you want to keep in
to the source state.
