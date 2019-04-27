# Frequently asked questions

* [What are the consequences of "bare" modifications to the target files? If my `.zshrc` is managed by chezmoi and I edit `~/.zshrc` without using `chezmoi edit`, what happens?](#what-are-the-consequences-of-bare-modifications-to-the-target-files-if-my-zshrc-is-managed-by-chezmoi-and-i-edit-zshrc-without-using-chezmoi-edit-what-happens)
* [How can I tell what dotfiles in my home directory aren't managed by chezmoi? Is there an easy way to have chezmoi manage a subset of them?](#how-can-i-tell-what-dotfiles-in-my-home-directory-arent-managed-by-chezmoi-is-there-an-easy-way-to-have-chezmoi-manage-a-subset-of-them)
* [If there's a mechanism in place for the above, is there also a way to tell chezmoi to ignore specific files or groups of files (e.g. by directory name or by glob)?](#if-theres-a-mechanism-in-place-for-the-above-is-there-also-a-way-to-tell-chezmoi-to-ignore-specific-files-or-groups-of-files-eg-by-directory-name-or-by-glob)
* [If the target already exists, but is "behind" the source, can chezmoi be configured to preserve the target version before replacing it with one derived from the source?](#if-the-target-already-exists-but-is-behind-the-source-can-chezmoi-be-configured-to-preserve-the-target-version-before-replacing-it-with-one-derived-from-the-source)
* [I've made changes to both the destination state and the source state that I want to keep. How can I keep them both?](#ive-made-changes-to-both-the-destination-state-and-the-source-state-that-i-want-to-keep-how-can-i-keep-them-both)
* [How do I tell chezmoi to always delete a file?](#how-do-i-tell-chezmoi-to-always-delete-a-file)
* [What other questions have been asked about chezmoi?](#what-other-questions-have-been-asked-about-chezmoi)
* [Where do I ask a question that isn't answered here?](#where-do-i-ask-a-question-that-isnt-answered-here)

## What are the consequences of "bare" modifications to the target files?  If my `.zshrc` is managed by chezmoi and I edit `~/.zshrc` without using `chezmoi edit`, what happens?

chezmoi will overwrite the file the next time you run `chezmoi apply`. Until you
run `chezmoi apply` your modified `~/.zshrc` will remain in place.

## How can I tell what dotfiles in my home directory aren't managed by chezmoi? Is there an easy way to have chezmoi manage a subset of them?

`chezmoi unmanaged` will list everything not managed by chezmoi. You can add
entire directories with `chezmoi add -r`.

## If there's a mechanism in place for the above, is there also a way to tell chezmoi to ignore specific files or groups of files (e.g. by directory name or by glob)?

By default, chezmoi ignores everything that you haven't explicitly `chezmoi
add`ed. If have files in your source directory that you don't want added to your
destination directory when you run `chezmoi apply` add them to a
`.chezmoiignore` file (which supports globs and is also a template).

## If the target already exists, but is "behind" the source, can chezmoi be configured to preserve the target version before replacing it with one derived from the source?

Yes. Run `chezmoi add` will update the source state with the target. To see
diffs of what would change, without actually changing anything, use `chezmoi
diff`.

## I've made changes to both the destination state and the source state that I want to keep. How can I keep them both?

`chezmoi merge` will open a merge tool to resolve differences between the source
state, target state, and destination state. Copy the changes you want to keep in
to the source state.

## How do I tell chezmoi to always delete a file?

chezmoi will delete files if their target state is empty, unless they have the
`empty` attribute set. Therefore an empty file in the source state without the
`empty` attribute will always be deleted.

Say you want chezmoi to always delete `~/.foo`, you can use the following
sequence of commands:

    rm -f ~/.foo
    touch ~/.foo
    chezmoi add --empty ~/.foo
    chezmoi chattr noempty ~/.foo

When you next run `chezmoi apply`, `~/.foo` will be deleted.

## What other questions have been asked about chezmoi?

See the [issues on
Github](https://github.com/twpayne/chezmoi/issues?utf8=%E2%9C%93&q=is%3Aissue+sort%3Aupdated-desc+label%3Asupport).

## Where do I ask a question that isn't answered here?

Please [open an issue on Github](https://github.com/twpayne/chezmoi/issues/new).