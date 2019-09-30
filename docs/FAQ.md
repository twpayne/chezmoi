# chezmoi Frequently Asked Questions

* [How can I quickly check for problems with chezmoi on my machine?](#how-can-i-quickly-check-for-problems-with-chezmoi-on-my-machine)
* [What are the consequences of "bare" modifications to the target files? If my `.zshrc` is managed by chezmoi and I edit `~/.zshrc` without using `chezmoi edit`, what happens?](#what-are-the-consequences-of-bare-modifications-to-the-target-files-if-my-zshrc-is-managed-by-chezmoi-and-i-edit-zshrc-without-using-chezmoi-edit-what-happens)
* [How can I tell what dotfiles in my home directory aren't managed by chezmoi? Is there an easy way to have chezmoi manage a subset of them?](#how-can-i-tell-what-dotfiles-in-my-home-directory-arent-managed-by-chezmoi-is-there-an-easy-way-to-have-chezmoi-manage-a-subset-of-them)
* [If there's a mechanism in place for the above, is there also a way to tell chezmoi to ignore specific files or groups of files (e.g. by directory name or by glob)?](#if-theres-a-mechanism-in-place-for-the-above-is-there-also-a-way-to-tell-chezmoi-to-ignore-specific-files-or-groups-of-files-eg-by-directory-name-or-by-glob)
* [If the target already exists, but is "behind" the source, can chezmoi be configured to preserve the target version before replacing it with one derived from the source?](#if-the-target-already-exists-but-is-behind-the-source-can-chezmoi-be-configured-to-preserve-the-target-version-before-replacing-it-with-one-derived-from-the-source)
* [I've made changes to both the destination state and the source state that I want to keep. How can I keep them both?](#ive-made-changes-to-both-the-destination-state-and-the-source-state-that-i-want-to-keep-how-can-i-keep-them-both)
* [gpg encryption fails. What could be wrong?](#gpg-encryption-fails-what-could-be-wrong)
* [What inspired chezmoi?](#what-inspired-chezmoi)
* [Can I use chezmoi outside my home directory?](#can-i-use-chezmoi-outside-my-home-directory)
* [Where does the name "chezmoi" come from?](#where-does-the-name-chezmoi-come-from)
* [What other questions have been asked about chezmoi?](#what-other-questions-have-been-asked-about-chezmoi)
* [Where do I ask a question that isn't answered here?](#where-do-i-ask-a-question-that-isnt-answered-here)

## How can I quickly check for problems with chezmoi on my machine?

Run:

    chezmoi doctor

Anything `ok` is fine, anything `warning` is only a problem if you want to use
the related feature, and anything `error` indicates a definite problem.

## What are the consequences of "bare" modifications to the target files? If my `.zshrc` is managed by chezmoi and I edit `~/.zshrc` without using `chezmoi edit`, what happens?

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

## gpg encryption fails. What could be wrong?

The `gpg.recipient` key should be ultimately trusted, otherwise encryption will
fail because gpg will prompt for input, which chezmoi does not handle. You can
check the trust level by running:

    gpg --export-ownertrust

The trust level for the recipient's key should be `6`. If it is not, you can
change the trust level by running:

    gpg --edit-key $recipient

Enter `trust` at the prompt and chose `5 = I trust ultimately`.

## What inspired chezmoi?

chezmoi was inspired by [Puppet](https://puppet.com/), but created because
Puppet is a slow overkill for managing your personal configuration files. The
focus of chezmoi will always be personal home directory management. If your
needs grow beyond that, switch to a whole system configuration management tool.

## Can I use chezmoi outside my home directory?

chezmoi, by default, operates on your home directory, but this can be overridden
with the `--destination` command line flag or by specifying `destDir` in your config
file. In theory, you could use chezmoi to manage any aspect of your filesystem.
That said, although you can do this, you probably shouldn't. Existing
configuration management tools like [Puppet](https://puppet.com/),
[Chef](https://www.chef.io/chef/), [Ansible](https://www.ansible.com/), and
[Salt](https://www.saltstack.com/) are much better suited to whole system
configuration management.

## Where does the name "chezmoi" come from?

"chezmoi" splits to "chez moi" and pronouced /ʃeɪ mwa/ (shay-moi) meaning "at my
house" in French. It's seven letters long, which is an appropriate length for a
command that is only run occasionally.

## What other questions have been asked about chezmoi?

See the [issues on
GitHub](https://github.com/twpayne/chezmoi/issues?utf8=%E2%9C%93&q=is%3Aissue+sort%3Aupdated-desc+label%3Asupport).

## Where do I ask a question that isn't answered here?

Please [open an issue on GitHub](https://github.com/twpayne/chezmoi/issues/new).
