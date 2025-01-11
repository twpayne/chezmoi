# Design

## Do I have to use `chezmoi edit` to edit my dotfiles?

No. `chezmoi edit` is a convenience command that has a couple of useful
features, but you don't have to use it.

You can also run `chezmoi cd` and then just edit the files in the source state
directly. After saving an edited file you can run `chezmoi diff` to check what
effect the changes would have, and run `chezmoi apply` if you're happy with
them. If there are inconsistencies that you want to keep, then `chezmoi
merge-all` will help you resolve any differences.

`chezmoi edit` provides the following useful features:

* The arguments to `chezmoi edit` are the files in their target location, so you
  don't have to think about source state attributes and your editor's syntax
  highlighting will work.

* If the dotfile is encrypted in the source state, then `chezmoi edit` will
  decrypt it to a private directory, open that file in your `$EDITOR`, and then
  re-encrypt the file when you quit your editor. This makes encryption
  transparent.

* With the `--diff` and `--apply` options you can see what would change and
  apply those changes without having to run `chezmoi diff` or `chezmoi apply`.

* If you have configured git auto commits or git auto pushes then `chezmoi edit`
  will create commits and push them for you.

If you chose to edit files in the source state and you're using VIM then
[`github.com/alker0/chezmoi.vim`](https://github.com/alker0/chezmoi.vim) gives
you syntax highlighting, however you edit your files. Besides using the
plugin, you can use modeline to tell VIM the correct filetype. For example,
put `# vim: filetype=zsh` at the top of `dot_zshrc`, and VIM will treat
`dot_zshrc` as zsh file.

## Why doesn't chezmoi use symlinks like GNU Stow?

Symlinks are first class citizens in chezmoi: chezmoi supports creating them,
updating them, removing them, and even more advanced features not found in other
dotfile managers like having the same symlink point to different targets on
different machines by using a template.

With chezmoi, you only use a symlink where you really need a symlink, in
contrast to some other dotfile managers (e.g. GNU Stow) which require the use
of symlinks as a layer of indirection between a dotfile's location (which can
be anywhere in your home directory) and a dotfile's content (which needs to be
in a centralized directory that you manage with version control). chezmoi
solves this problem in a different way.

Instead of using a symlink to redirect from the dotfile's location to the
centralized directory, chezmoi generates the dotfile as a regular file in its
final location from the contents of the centralized directory. This approach
allows chezmoi to provide features that are not possible when using symlinks,
for example having files that are encrypted, executable, private, or templates.

There is nothing special about dotfiles managed by chezmoi whereas dotfiles
managed with GNU Stow are special because they're actually symlinks to somewhere
else.

The only advantage to using GNU Stow-style symlinks is that changes that you
make to the dotfile's contents in the centralized directory are immediately
visible whenever you save them, whereas chezmoi currently requires you to pass
the `--watch` flag to `chezmoi edit` or set `edit.watch` to `true` in your
configuration file.

If you really want to use symlinks, then chezmoi provides a [symlink
mode](../../reference/target-types.md#symlink-mode) which uses symlinks where
possible. This configures chezmoi to work like GNU Stow and have it create a set
of symlinks back to a central directory, but this currently requires a bit of
manual work (as described in
[#167](https://github.com/twpayne/chezmoi/issues/167)). chezmoi might get some
automation to help (see [#886](https://github.com/twpayne/chezmoi/issues/886)
for example) but it does need some convincing use cases that demonstrate that a
symlink from a dotfile's location to its contents in a central directory is
better than just having the correct dotfile contents.

## What are the limitations of chezmoi's symlink mode?

In symlink mode chezmoi replaces targets with symlinks to the source directory
if the target is a regular file and is not encrypted, executable, private,
or a template.

Symlinks cannot be used for encrypted files because the source state contains
the ciphertext, not the plaintext.

Symlinks cannot be used for executable files as the executable bit would need
to be set on the file in the source directory and chezmoi uses only regular
files and directories in its source state for portability across operating
systems. This may change in the future.

Symlinks cannot be used for private files because git does not persist group
and world permission bits.

Symlinks cannot be used for templated files because the source state contains
the template, not the result of executing the template.

Symlinks cannot be used for entire directories because of chezmoi's use of
attributes in the filename mangles entries in the directory, directories might
have the `exact_` attribute and contain empty files, and the directory's
entries might not be usable with symlinks.

In symlink mode, running `chezmoi add` does not immediately replace the targets
with a symlink. You must run `chezmoi apply` to create the symlinks.

## Why does chezmoi use weird filenames?

There are a number of criticisms of how chezmoi uses filenames:

1. The long source file names are weird and verbose.
2. Not all possible file permissions can be represented.
3. Everything is in a single directory, which can end up containing many
   entries.

chezmoi's decision to store metadata in filenames is a deliberate, practical,
compromise.

Firstly, almost all programs store metadata in filenames: the filename's
extension. chezmoi extends the filename to storing metadata in attributes in the
filename's prefix as well.

The `dot_` attribute makes it transparent which dotfiles are managed by chezmoi
and which files are ignored by chezmoi. chezmoi ignores all files and
directories that start with `.` so no special whitelists are needed for version
control systems and their control files (e.g. `.git` and `.gitignore`).

chezmoi needs per-file metadata to know how to interpret the source file's
contents, for example to know when the source file is a template or if the
file's contents are encrypted. By storing this metadata in the filename, the
metadata is unambiguously associated with a single file and adding, updating,
or removing a single file touches only a single file in the source state.
Changes to the metadata (e.g. `chezmoi chattr +template $TARGET`) are simple
file renames and isolated to the affected file.

If chezmoi were to, say, use a common configuration file listing which files
were templates and/or encrypted, then changes to any file would require updates
to the common configuration file. Automating updates to configuration files
requires a round trip (read config file, update config, write config) and it is
not always possible preserve comments and formatting.

chezmoi's attributes of `executable_`, `private_`, and `readonly_` allow the
file permissions `0o644`, `0o755`, `0o600`, `0o700`, `0o444`, `0o555`, `0o400`,
and `0o500` to be represented. Directories can only have permissions `0o755`,
`0o700`, or `0o500`. In practice, these cover all permissions typically used
for dotfiles. If this does cause a genuine problem for you, please [open an
issue on GitHub](https://github.com/twpayne/chezmoi/issues/new/choose).

File permissions and modes like `executable_`, `private_`, `readonly_`, and
`symlink_` could also be stored in the filesystem, rather than in the filename.
However, this requires the permissions to be preserved and handled by the
underlying version control system and filesystem. chezmoi provides first-class
support for Windows, where the `executable_` and `private_` attributes have no
direct equivalents and symbolic links are not always permitted. By using
regular files and directories, chezmoi avoids variations in the operating
system, version control system, and filesystem making it both more robust and
more portable.

chezmoi uses a 1:1 mapping between entries in the source state and entries in
the target state. This mapping is bi-directional and unambiguous.

However, this also means that dotfiles that in the same directory in the target
state must be in the same directory in the source state. In particular, every
entry managed by chezmoi in the root of your home directory has a corresponding
entry in the root of your source directory, which can mean that you end up with
a lot of entries in the root of your source directory. This can be mitigated by
using `.chezmoiroot` file.

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

## Can chezmoi support multiple sources or multiple source states?

With some dotfile managers, dotfiles can be distributed across multiple
directories or even multiple repos. For example, the user might have one
directory per application, or separate repos for home and work configurations,
or even separate git submodules for different applications. These can be
considered multiple sources of truth for the target state. This, however, comes
with complications:

1. Multiple sources of truth complicate the user interface. When running
   `chezmoi add $FILE`, which source should `$FILE` be added to?

2. Multiple sources of truth do not compose easily if target files overlap. For
   example, if you have two sources, both of which need to set an environment
   variable in `.bashrc`, how do you handle this when both, only one, or
   neither source might be activated? What if the sources are mutually
   exclusive, e.g. if the VIM source and the Emacs source both want to set the
   `$EDITOR` environment variable?

3. Multiple sources of truth are not always independent. Related to the
   previous point, consider a source that adds an applications's configuration
   files and shell completions. Should the shell completions be part of the
   applications's source or of the shell's source?

chezmoi instead makes the opinionated choice to use a single source of truth,
i.e. a single branch in a single git repo. Using a single source of truth
avoids the inherent complexity and ambiguity of multiple sources.

chezmoi provides mechanisms like templates (for minor differences),
`.chezmoiignore` (for controlling the presence or otherwise of complete files
and directories), and password manager integration (so secrets never need to be
stored in a repo) handle machine-to-machine differences. Externals make it easy
to pull in dotfiles from third-party sources.

That said, if you are keen to use multiple sources of truth with chezmoi, you
have a number of options with some scripting around chezmoi.

Firstly, you can run `chezmoi apply` with different arguments to the `--config`
and `--source` flags which will apply to the same destination. So that you only
have to type one command you can wrap this in a shell function, for example:

```bash
chezmoi-apply() {
    chezmoi apply --config ~/.config/chezmoi-home/chezmoi.toml \
                  --source ~/.local/share/chezmoi-home && \
    chezmoi apply --config ~/.config/chezmoi-work/chezmoi.toml \
                  --source ~/.local/share/chezmoi-work
}
```

If you want to generate multiple configuration files with `chezmoi init` then
you will need the `--config-path` flag. For more advanced use, use the
`--destination`, `--cache`, and `--persistent-state` flags.

Secondly, you can assemble a single source state from multiple sources and then
use `chezmoi apply`. For example, if you have multiple source states in
subdirectories of `~/.dotfiles`:

```bash
#!/bin/bash

# create a combined source state in a temporary directory
combined_source="$(mktemp -d)"

# remove the temporary source state on exit
trap 'rm -rf -- "${combined_source}"' INT TERM

# copy files from multiple sources into the temporary source state
for source in $HOME/.dotfiles/*; do
    cp -r "${source}"/* "${combined_source}"
done

# apply the temporary source state
chezmoi apply --source "${combined_source}"
```

Thirdly, you can use a `run_` script to invoke a second instance of chezmoi,
[as used by
@felipecrs](https://github.com/felipecrs/dotfiles/blob/8a7840efdeff1a45069f47e5b2e558dc9812712d/home/.chezmoiscripts/run_after_20-run-rootmoi.sh.tmpl).

## Why does `chezmoi cd` spawn a shell instead of just changing directory?

`chezmoi cd` spawns a shell because it is not possible for a program to change
the working directory of its parent process. You can add a shell function instead:

```bash
chezmoi-cd() {
    cd $(chezmoi source-path)
}
```

Typing `chezmoi-cd` will then change the directory of your current shell to
chezmoi's source directory.

## Why are the `prompt*` functions only available in config file templates?

chezmoi regularly needs to execute templates to determine the target contents
of files. For example, templates are executed for the `apply`, `diff`, and
`status` commands, amongst many others. Having to interactively respond each
time would quickly become tiresome. Therefore, chezmoi only provides these
functions when generating a config file from a config file template (e.g. when
you run `chezmoi init` or `chezmoi --init apply`).

## Why not use Ansible/Chef/Puppet/Salt, or similar to manage my dotfiles instead?

Whole system management tools are more than capable of managing your dotfiles,
but they are large systems that entail several disadvantages. Compared to whole
system management tools, chezmoi offers:

* Small, focused feature set designed for dotfiles. There's simply less to learn
  with chezmoi compared to whole system management tools.

* Easy installation and execution on every platform, without root access.
  Installing chezmoi requires only copying a single binary file with no external
  dependencies. Executing chezmoi just involves running the binary. In contrast,
  installing and running a whole system management tool typically requires
  installing a scripting language runtime, several packages, and running a
  system service, all typically requiring root access.

chezmoi's focus and simple installation means that it runs almost everywhere:
from tiny ARM-based Linux systems to Windows desktops, from inside lightweight
containers to FreeBSD-based virtual machines in the cloud.

## Can I use chezmoi to manage files outside my home directory?

In practice, yes, you can, but this usage is strongly discouraged beyond using
your system's package manager to install the packages you need.

chezmoi is designed to operate on your home directory, and is explicitly not a
full system configuration management tool. That said, there are some ways to
have chezmoi manage a few files outside your home directory.

chezmoi's scripts can execute arbitrary commands, so you can use a `run_` script
that is run every time you run `chezmoi apply`, to, for example:

* Make the target file outside your home directory a symlink to a file managed
  by chezmoi in your home directory.

* Copy a file managed by chezmoi inside your home directory to the target file.

* Execute a template with `chezmoi execute-template --output=$FILENAME
  template` where `$FILENAME` is outside the target directory.

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

## What inspired chezmoi?

chezmoi was inspired by [Puppet](https://puppet.com/), but was created because
Puppet is an overkill for managing your personal configuration files. The focus
of chezmoi will always be personal home directory management. If your needs
grow beyond that, switch to a whole system configuration management tool.

## Where does the name "chezmoi" come from?

"chezmoi" splits to "chez moi" and pronounced /ʃeɪ mwa/ (shay-mwa) meaning "at
my house" in French. It's seven letters long, which is an appropriate length for
a command that is only run occasionally. If you prefer a shorter command, add an
alias to your shell configuration, for example:

```sh
alias cz=chezmoi
```
