# chezmoi Frequently Asked Questions

<!--- toc --->
* [How can I quickly check for problems with chezmoi on my machine?](#how-can-i-quickly-check-for-problems-with-chezmoi-on-my-machine)
* [What are the consequences of "bare" modifications to the target files? If my `.zshrc` is managed by chezmoi and I edit `~/.zshrc` without using `chezmoi edit`, what happens?](#what-are-the-consequences-of-bare-modifications-to-the-target-files-if-my-zshrc-is-managed-by-chezmoi-and-i-edit-zshrc-without-using-chezmoi-edit-what-happens)
* [How can I tell what dotfiles in my home directory aren't managed by chezmoi? Is there an easy way to have chezmoi manage a subset of them?](#how-can-i-tell-what-dotfiles-in-my-home-directory-arent-managed-by-chezmoi-is-there-an-easy-way-to-have-chezmoi-manage-a-subset-of-them)
* [How can I tell what dotfiles in my home directory are currently managed by chezmoi?](#how-can-i-tell-what-dotfiles-in-my-home-directory-are-currently-managed-by-chezmoi)
* [If there's a mechanism in place for the above, is there also a way to tell chezmoi to ignore specific files or groups of files (e.g. by directory name or by glob)?](#if-theres-a-mechanism-in-place-for-the-above-is-there-also-a-way-to-tell-chezmoi-to-ignore-specific-files-or-groups-of-files-eg-by-directory-name-or-by-glob)
* [If the target already exists, but is "behind" the source, can chezmoi be configured to preserve the target version before replacing it with one derived from the source?](#if-the-target-already-exists-but-is-behind-the-source-can-chezmoi-be-configured-to-preserve-the-target-version-before-replacing-it-with-one-derived-from-the-source)
* [Once I've made a change to the source directory, how do I commit it?](#once-ive-made-a-change-to-the-source-directory-how-do-i-commit-it)
* [How do I only run a script when a file has changed?](#how-do-i-only-run-a-script-when-a-file-has-changed)
* [I've made changes to both the destination state and the source state that I want to keep. How can I keep them both?](#ive-made-changes-to-both-the-destination-state-and-the-source-state-that-i-want-to-keep-how-can-i-keep-them-both)
* [Why does chezmoi convert all my template variables to lowercase?](#why-does-chezmoi-convert-all-my-template-variables-to-lowercase)
* [chezmoi makes `~/.ssh/config` group writeable. How do I stop this?](#chezmoi-makes-sshconfig-group-writeable-how-do-i-stop-this)
* [Why doesn't chezmoi use symlinks like GNU Stow?](#why-doesnt-chezmoi-use-symlinks-like-gnu-stow)
* [Do I have to use `chezmoi edit` to edit my dotfiles?](#do-i-have-to-use-chezmoi-edit-to-edit-my-dotfiles)
* [Can I change how chezmoi's source state is represented on disk?](#can-i-change-how-chezmois-source-state-is-represented-on-disk)
* [gpg encryption fails. What could be wrong?](#gpg-encryption-fails-what-could-be-wrong)
* [chezmoi reports `chezmoi: user: lookup userid NNNNN: input/output error`](#chezmoi-reports-chezmoi-user-lookup-userid-nnnnn-inputoutput-error)
* [chezmoi reports `chemzoi: timeout` or `chemzoi: timeout obtaining persistent state lock`](#chezmoi-reports-chemzoi-timeout-or-chemzoi-timeout-obtaining-persistent-state-lock)
* [I'm getting errors trying to build chezmoi from source](#im-getting-errors-trying-to-build-chezmoi-from-source)
* [What inspired chezmoi?](#what-inspired-chezmoi)
* [Why not use Ansible/Chef/Puppet/Salt, or similar to manage my dotfiles instead?](#why-not-use-ansiblechefpuppetsalt-or-similar-to-manage-my-dotfiles-instead)
* [Can I use chezmoi to manage files outside my home directory?](#can-i-use-chezmoi-to-manage-files-outside-my-home-directory)
* [Where does the name "chezmoi" come from?](#where-does-the-name-chezmoi-come-from)
* [What other questions have been asked about chezmoi?](#what-other-questions-have-been-asked-about-chezmoi)
* [Where do I ask a question that isn't answered here?](#where-do-i-ask-a-question-that-isnt-answered-here)
* [I like chezmoi. How do I say thanks?](#i-like-chezmoi-how-do-i-say-thanks)

## How can I quickly check for problems with chezmoi on my machine?

Run:

    chezmoi doctor

Anything `ok` is fine, anything `warning` is only a problem if you want to use
the related feature, and anything `error` indicates a definite problem.

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
manual](https://github.com/twpayne/chezmoi/blob/master/docs/REFERENCE.md#chezmoiignore).

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
  guide](https://github.com/twpayne/chezmoi/blob/master/docs/HOWTO.md#automatically-commit-and-push-changes-to-your-repo).

## How do I only run a script when a file has changed?

A common example of this is that you're using [Homebrew](https://brew.sh/) and
have `.Brewfile` listing all the packages that you want installed and only want
to run `brew bundle --global` when the contents of `.Brewfile` have changed.

chezmoi has two types of scripts: scripts that run every time, and scripts that
only run when their contents change. chezmoi does not have a mechanism to run a
script when an arbitrary file has changed, but there are some ways to achieve
the desired behavior:

1. Have the script create `.Brewfile` instead of chezmoi, e.g. in your
   `run_once_install-packages`:

   ```sh
   #!/bin/sh

   cat > $HOME/.Brewfile <<EOF
   brew "imagemagick"
   brew "openssl"
   EOF

   brew bundle --global
   ```

2. Don't use `.Brewfile`, and instead install the packages explicitly in
   `run_once_install-packages`:

   ```sh
   #!/bin/sh

   brew install imagemagick || true
   brew install openssl || true
   ```

   The `|| true` is necessary because `brew install` exits with failure if the
   package is already installed.

3. Use a script that runs every time (not just once) and rely on `brew bundle
   --global` being idempotent.

4. Use a script that runs every time, records a checksum of `.Brewfile` in
   another file, and only runs `brew bundle --global` if the checksum has
   changed, and updates the recorded checksum after.

## I've made changes to both the destination state and the source state that I want to keep. How can I keep them both?

`chezmoi merge` will open a merge tool to resolve differences between the source
state, target state, and destination state. Copy the changes you want to keep in
to the source state.

## Why does chezmoi convert all my template variables to lowercase?

This is due to a feature in
[`github.com/spf13/viper`](https://github.com/spf13/viper), the library that
chezmoi uses to read its configuration file. For more information see [this
GitHub issue](https://github.com/twpayne/chezmoi/issues/463).

## chezmoi makes `~/.ssh/config` group writeable. How do I stop this?

By default, chezmoi uses your system's umask when creating files. On most
systems the default umask is `022` but some systems use `002`, which means
that files and directories are group writeable by default.

You can override this for chezmoi by setting the `umask` configuration variable
in your configuration file, for example:

    umask = 0o022

Note that this will apply to all files and directories that chezmoi manages and
will ensure that none of them are group writeable. It is not currently possible
to control group write permissions for individual files or directories. Please
[open an issue on
GitHub](https://github.com/twpayne/chezmoi/issues/new?assignees=&labels=enhancement&template=02_feature_request.md&title=)
if you need this.

## Why doesn't chezmoi use symlinks like GNU Stow?

Symlinks are first class citizens in chezmoi: chezmoi supports creating them,
updating them, removing them, and even more advanced features not found
elsewhere like having the same symlink point to different targets on different
machines by using templates.

With chezmoi, you only use a symlink where you really want a symlink, in
contrast to some other dotfile managers (e.g. GNU Stow) which require the use of
symlinks as a layer of indirection between a dotfile's location (which can be
anywhere in your home directory) and a dotfile's content (which needs to be in a
centralized directory that you manage with version control). chezmoi solves this
problem in a different way.

Instead of using a symlink to redirect from the dotfile's location to the
centralized directory, chezmoi generates the dotfile in its final location from
the contents of the centralized directory. Not only is no symlink is needed,
this has the advantages that chezmoi is better able to cope with differences
from machine to machine (as a dotfile's contents can be unique to that machine)
and the dotfiles that chezmoi creates are just regular files. There's nothing
special about dotfiles managed by chezmoi, whereas dotfiles managed with GNU
Stow are special because they're actually symlinks to somewhere else.

The only advantage to using GNU Stow-style symlinks is that changes that you
make to the dotfile's contents in the centralized directory are immediately
visible, whereas chezmoi currently requires you to run `chezmoi apply` or
`chezmoi edit --apply`. chezmoi will likely get an alternative solution to this
too, see [#752](https://github.com/twpayne/chezmoi/issues/752).

You can configure chezmoi to work like GNU Stow and have it create a set of
symlinks back to a central directory, but this currently requires a bit of
manual work (as described in
[#167](https://github.com/twpayne/chezmoi/issues/167)). chezmoi might get some
automation to help (see [#886](https://github.com/twpayne/chezmoi/issues/886)
for example) but it does need some convincing use cases that demonstrate that a
symlink from a dotfile's location to its contents in a central directory is
better than just having the correct dotfile contents.

## Do I have to use `chezmoi edit` to edit my dotfiles?

No. `chezmoi edit` is a convenience command that has a couple of useful
features, but you don't have to use it. You can also run `chezmoi cd` and then
just edit the files in the source state directly. After saving an edited file
you can run `chezmoi diff` to check what effect the changes would have, and run
`chezmoi apply` if you're happy with them.

`chezmoi edit` provides the following useful features:
* It opens the correct file in the source state for you, so you don't have to
  know anything about source state attributes.
* If the dotfile is encrypted in the source state, then `chezmoi edit` will
  decrypt it to a private directory, open that file in your `$EDITOR`, and then
  re-encrypt the file when you quit your editor. That makes encryption more
  transparent to the user. With the `--diff` and `--apply` options you can see
  what would change and apply those changes without having to run `chezmoi diff`
  or `chezmoi apply`. Note also that the arguments to `chezmoi edit` are the
  files in their target location.

## Can I change how chezmoi's source state is represented on disk?

There are a number of criticisms of how chezmoi's source state is represented on
disk:

1. The source file naming system cannot handle all possible filenames.
2. Not all possible file permissions can be represented.
3. The long source file names are weird and verbose.
4. Everything is in a single directory, which can end up containing many entries.

chezmoi's source state representation is a deliberate, practical compromise.

Certain target filenames, for example `~/dot_example`, are incompatible with
chezmoi's
[attributes](https://github.com/twpayne/chezmoi/blob/master/docs/REFERENCE.md#source-state-attributes)
used in the source state. In practice, dotfile filenames are unlikely to
conflict with chezmoi's attributes. If this does cause a genuine problem for
you, please [open an issue on
GitHub](https://github.com/twpayne/chezmoi/issues/new/choose).

The `dot_` attribute makes it transparent which dotfiles are managed by chezmoi
and which files are ignored by chezmoi. chezmoi ignores all files and
directories that start with `.` so no special whitelists are needed for version
control systems and their control files (e.g. `.git` and `.gitignore`).

chezmoi needs per-file metadata to know how to interpret the source file's
contents, for example to know when the source file is a template or if the
file's contents are encrypted. By storing this metadata in the filename, the
metadata is unambiguously associated with a single file and adding, updating, or
removing a single file touches only a single file in the source state. Changes
to the metadata (e.g. `chezmoi chattr +template *target*`) are simple file
renames and isolated to the affected file.

If chezmoi were to, say, use a common configuration file listing which files
were templates and/or encrypted, then changes to any file would require updates
to the common configuration file. Automating updates to configuration files
requires a round trip (read config file, update config, write config) and it is
not always possible preserve comments and formatting.

chezmoi's attributes of `executable_` and `private_` only allow a the file
permissions `0o644`, `0o755`, `0o600`, and `0o700` to be represented.
Directories can only have permissions `0o755` or `0o700`. In practice, these
cover all permissions typically used for dotfiles. If this does cause a genuine
problem for you, please [open an issue on
GitHub](https://github.com/twpayne/chezmoi/issues/new/choose).

File permissions and modes like `executable_`, `private_`, and `symlink_` could
also be stored in the filesystem, rather than in the filename. However, this
requires the permissions to be preserved and handled by the underlying version
control system and filesystem. chezmoi provides first-class support for Windows,
where the `executable_` and `private_` attributes have no direct equivalents and
symbolic links are not always permitted. By using regular files and directories,
chezmoi avoids variations in the operating system, version control system, and
filesystem making it both more robust and more portable.

chezmoi uses a 1:1 mapping between entries in the source state and entries in
the target state. This mapping is bi-directional and unambiguous.

However, this also means that dotfiles that in the same directory in the target
state must be in the same directory in the source state. In particular, every
entry managed by chezmoi in the root of your home directory has a corresponding
entry in the root of your source directory, which can mean that you end up with
a lot of entries in the root of your source directory.

If chezmoi were to permit, say, multiple separate source directories (so you
could, say, put `dot_bashrc` in a `bash/` subdirectory, and `dot_vimrc` in a
`vim/` subdirectory, but have `chezmoi apply` map these to `~/.bashrc` and
`~/.vimrc` in the root of your home directory) then the mapping between source
and target states is no longer bidirectional nor unambiguous, which
significantly increases complexity and requires more user interaction. For
example, if both `bash/dot_bashrc` and `vim/dot_bashrc` exist, what should be
the contents of `~/.bashrc`? If you run `chezmoi add ~/.zshrc`, should
`dot_zshrc` be stored in the source `bash/` directory, the source `vim/`
directory, or somewhere else? How does the user communicate their preferences?

chezmoi has many users and any changes to the source state representation must
be backwards-compatible.

In summary, chezmoi's source state representation is a compromise with both
advantages and disadvantages. Changes to the representation will be considered,
but must meet the following criteria, in order of importance:

1. Be fully backwards-compatible for existing users.
2. Fix a genuine problem encountered in practice.
3. Be independent of the underlying operating system, version control system,
   and filesystem.
4. Not add significant extra complexity to the user interface or underlying
   implementation.

## gpg encryption fails. What could be wrong?

The `gpg.recipient` key should be ultimately trusted, otherwise encryption will
fail because gpg will prompt for input, which chezmoi does not handle. You can
check the trust level by running:

    gpg --export-ownertrust

The trust level for the recipient's key should be `6`. If it is not, you can
change the trust level by running:

    gpg --edit-key $recipient

Enter `trust` at the prompt and chose `5 = I trust ultimately`.

## chezmoi reports `chezmoi: user: lookup userid NNNNN: input/output error`

This is likely because the chezmoi binary you are using was statically compiled
with [musl](https://musl.libc.org/) and the machine you are running on uses
LDAP or NIS.

The immediate fix is to use a package built for your distribution (e.g a `.deb`
or `.rpm`) which is linked against glibc and includes LDAP/NIS support instead
of the statically-compiled binary.

If the problem still persists, then please [open an issue on
GitHub](https://github.com/twpayne/chezmoi/issues/new/choose).

## chezmoi reports `chemzoi: timeout` or `chemzoi: timeout obtaining persistent state lock`

chezmoi will report this when it is unable to lock its persistent state
(`~/.config/chezmoi/chezmoistate.boltdb`), typically because another instance of
chezmoi is currently running and holding the lock.

This can happen, for example, if you have a `run_` script that invokes
`chezmoi`, or are running chezmoi in another window.

Under the hood, chezmoi uses [bbolt](https://github.com/etcd-io/bbolt) which
permits multiple simultaneous readers, but only one writer (with no readers).

Commands that take a write lock include `add`, `apply`, `edit`, `import`,
`init`, `state`, and `update`. Commands that take a read lock include `diff`,
`status`, and `verify`.

## I'm getting errors trying to build chezmoi from source

chezmoi requires Go version 1.16 or later. You can check the version of Go with:

    go version

For more details on building chezmoi, see the [Contributing
Guide]([CONTRIBUTING.md](https://github.com/twpayne/chezmoi/blob/master/docs/CONTRIBUTING.md)).

## What inspired chezmoi?

chezmoi was inspired by [Puppet](https://puppet.com/), but was created because
Puppet is an overkill for managing your personal configuration files. The focus
of chezmoi will always be personal home directory management. If your needs grow
beyond that, switch to a whole system configuration management tool.

## Why not use Ansible/Chef/Puppet/Salt, or similar to manage my dotfiles instead?

Whole system management tools are more than capable of managing your dotfiles,
but are large systems that entail several disadvantages. Compared to whole
system management tools, chezmoi offers:

* Small, focused feature set designed for dotfiles. There's simply less to learn
  with chezmoi compared to whole system management tools.
* Easy installation and execution on every platform, without root access.
  Installing chezmoi requires only copying a single binary file with no external
  dependencies. Executing chezmoi just involves running the binary. In contrast,
  installing and running a whole system management tools typically requires
  installing a scripting language runtime, several packages, and running a
  system service, all typically requiring root access.

chezmoi's focus and simple installation means that it runs almost everywhere:
from tiny ARM-based Linux systems to Windows desktops, from inside lightweight
containers to FreeBSD-based virtual machines in the cloud.

## Can I use chezmoi to manage files outside my home directory?

In practice, yes, you can, but this is strongly discouraged beyond using your
system's package manager to install the packages you need.

chezmoi is designed to operate on your home directory, and is explicitly not a
full system configuration management tool. That said, there are some ways to
have chezmoi manage a few files outside your home directory.

chezmoi's scripts can execute arbitrary commands, so you can use a `run_` script
that is run every time you run `chezmoi apply`, to, for example:

* Make the target file outside your home directory a symlink to a file managed
  by chezmoi in your home directory.
* Copy a file managed by chezmoi inside your home directory to the target file.
* Execute a template with `chezmoi execute-template --output=filename template`
  where `filename` is outside the target directory.

chezmoi executes all scripts as the user executing chezmoi, so you may need to
add extra privilege elevation commands like `sudo` or `PowerShell start -verb
runas -wait` to your script.

chezmoi, by default, operates on your home directory but this can be overridden
with the `--destination` command line flag or by specifying `destDir` in your
config file, and could even be the root directory (`/` or `C:\`). This allows
you, in theory, to use chezmoi to manage any file in your filesystem, but this
usage is extremely strongly discouraged.

If your needs extend beyond modifying a handful of files outside your target
system, then existing configuration management tools like
[Puppet](https://puppet.com/), [Chef](https://chef.io/),
[Ansible](https://www.ansible.com/), and [Salt](https://www.saltstack.com/) are
much better suited - and of course can be called from a chezmoi `run_` script.
Put your Puppet Manifests, Chef Recipes, Ansible Modules, and Salt Modules in a
directory ignored by `.chezmoiignore` so they do not pollute your home
directory.

## Where does the name "chezmoi" come from?

"chezmoi" splits to "chez moi" and pronounced /ʃeɪ mwa/ (shay-moi) meaning "at
my house" in French. It's seven letters long, which is an appropriate length for
a command that is only run occasionally.

## What other questions have been asked about chezmoi?

See the [issues on
GitHub](https://github.com/twpayne/chezmoi/issues?utf8=%E2%9C%93&q=is%3Aissue+sort%3Aupdated-desc+label%3Asupport).

## Where do I ask a question that isn't answered here?

Please [open an issue on GitHub](https://github.com/twpayne/chezmoi/issues/new/choose).

## I like chezmoi. How do I say thanks?

Thank you! chezmoi was written to scratch a personal itch, and I'm very happy
that it's useful to you. Please give [chezmoi a star on
GitHub](https://github.com/twpayne/chezmoi/stargazers), and if you're happy to
share your public dotfile repo then [tag it with
`chezmoi`](https://github.com/topics/chezmoi?o=desc&s=updated).

If you write an article or give a talk on chezmoi please inform the author (e.g.
by [opening an issue](https://github.com/twpayne/chezmoi/issues/new/choose)) so
it can be added to chezmoi's [media
page](https://github.com/twpayne/chezmoi/blob/master/docs/MEDIA.md).

[Contributions are very
welcome](https://github.com/twpayne/chezmoi/blob/master/docs/CONTRIBUTING.md)
and every [bug report, support request, and feature
request](https://github.com/twpayne/chezmoi/issues/new/choose) helps make
chezmoi better. Thank you :)
