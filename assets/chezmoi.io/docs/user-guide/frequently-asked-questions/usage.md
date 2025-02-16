# Usage

## How do I edit my dotfiles with chezmoi?

There are five popular approaches:

1. Use `chezmoi edit $FILE`. This will open the source file for `$FILE` in your
   editor, including opening the template if the file is templated and
   transparently decrypting and re-encrypting it if it is encrypted. For extra
   ease, use `chezmoi edit --apply $FILE` to apply the changes when you quit
   your editor, and `chezmoi edit --watch $FILE` to apply the changes whenever
   you save the file.

2. Use `chezmoi cd` and edit the files in the source directory directly. Run
   `chezmoi diff` to see what changes would be made, and `chezmoi apply` to make
   the changes.

3. If your editor supports opening directories, run `chezmoi edit` with no
   arguments to open the source directory.

4. Edit the file in your home directory, and then either re-add it by running
   `chezmoi add $FILE` or `chezmoi re-add`.

5. Edit the file in your home directory, and then merge your changes with source
   state by running `chezmoi merge $FILE`.

    !!! note

        `re-add` doesn't work with templates.

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
machine. The full usage and syntax is described in the
[reference manual][ignore].

## If the target already exists, but is "behind" the source, can chezmoi be configured to preserve the target version before replacing it with one derived from the source?

Yes. Running `chezmoi add` will update the source state with the target. To see
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
  source state, as [described in the how-to guide][auto-commit].

## I've made changes to both the destination state and the source state that I want to keep. How can I keep them both?

`chezmoi merge` will open a merge tool to resolve differences between the source
state, target state, and destination state. Copy the changes you want to keep in
to the source state.

## Can I use chezmoi to manage my shell history across multiple machines?

No. Every change in a file managed by chezmoi requires an explicit command to
record it (e.g. `chezmoi add`) or apply it somewhere else (e.g. `chezmoi
update`), and is recorded as a commit in your dotfiles repository. Creating a
commit every time a command is entered would quickly become cumbersome. This
makes chezmoi unsuitable for sharing changes to rapidly-changing files like
shell histories.

Instead, consider using a dedicated tool for sharing shell history across
multiple machines, like [`atuin`][atuin]. You can use chezmoi to install and
configure atuin.

## How do I install pre-requisites for templates?

If you have a template that depends on some other tool, like `curl`, you may need
to install it before chezmoi renders the template.

To do so, use a `run_before` script that is **not** a template. Something like:

```bash title="run_before_00-install-pre-requisites.sh"
#!/bin/bash

set -eu

# Install curl if it's not already installed
if ! command -v curl >/dev/null; then
  sudo apt update
  sudo apt install -y curl
fi
```

chezmoi will make sure to execute it before templating other files.

!!! tip

    You can [use `scriptEnv` to inject data into your scripts through environment
    variables][scriptenv].

## How do I write a literal `{{` or `}}` in a template?

`{{` and `}}` are chezmoi's default template delimiters, and so need escaping, for example:

```text
{{ "{{" }}
{{ "}}" }}
```

results in

```text
{{
}}
```

For longer tokens containing a `{{` and a `}}` you can use a longer literal, for example:

```text
{{ "{{ .Target }}" }}
```

results in

```text
{{ .Target }}
```

## How do I run a script when a `git-repo` external changes?

Use a `run_onchange_after_*.tmpl` script that includes the HEAD commit. For example,
if `~/.emacs.d` is a `git-repo` external, then create:

```text title="~/.local/share/chezmoi/run_onchange_after_emacs.d.tmpl"
#!/bin/sh

# {{ output "git" "-C" (joinPath .chezmoi.homeDir ".emacs.d") "rev-parse" "HEAD" }}
echo "~/emacs.d updated"
```

## How do I run a script periodically?

Use a `run_once_*.tmpl` script that includes the current time truncated to a
suitable unit. For example, to run a script daily:

```text title="~/.local/share/chezmoi/run_once_daily.tmpl"
#!/bin/sh

# {{ now | date "2006-01-02" }}
echo "new day"
```

For weekly, use the week number from the output of `date`, for example:

```text title="~/.local/share/chezmoi/run_once_weekly.tmpl"
#!/bin/sh

# {{ output "date" "+%V" | trim }}
echo "new week"
```

Or, approximate the week number with template functions:

```text title="~/.local/share/chezmoi/run_once_weekly.tmpl"
#!/bin/sh

# {{ div now.YearDay 7 }}
echo "new week"
```

## How do I enable shell completions?

chezmoi includes shell completions for [`bash`][bash], [`fish`][fish],
[`powershell`][powershell], and [`zsh`][zsh]. If you have installed chezmoi via
your package manager then the shell completion should already be installed. For
PowerShell, you need to manually add the completion script to your profile.
Please [open an issue][choose] if this is not working correctly.

chezmoi provides a [`completion`][completion-cmd] command and
a [`completion`][completion-fun] template function which return the shell
completions for the given shell. These can be used either as a one-off or as
part of your dotfiles repo. The details of how to use these depend on your
shell.

## How do I use tools that I installed with Flatpak?

Command line programs installed with [Flatpak][flatpak] cannot be run directly.
Instead, they must be run with `flatpak run`. This can either be added by using
a wrapper script or by configuring chezmoi to invoke `flatpak run` with the
correct arguments directly. Wrapper scripts are recommended, as they work with
the [`doctor` command][doctor].

### Use a wrapper script

Create a wrapper script with the exact same name as the command that invokes
`flatpak run` and passes all arguments to the wrapped command.

For example, to wrap KeePassXC installed with Flatpak, create the script:

```bash title="keepassxc-cli"
#!/bin/bash

flatpak run --command=keepassxc-cli org.keepassxc.KeePassXC -- "$@"
```

Note that the script is called `keepassxc-cli` without any `.sh` extension, so
it has the exact same name as the `keepassxc-cli` command that chezmoi invokes
by default. Ensure that this script is in your path and is executable.

### Configure chezmoi to invoke `flatpak run`

For tools that chezmoi invokes with `.command` and `.args` configuration
variables, you can configure chezmoi to invoke `flatpak` directly with the
correct arguments.

For example, to use VSCodium installed with Flatpak as your diff command, add
the following to your config file:

```toml title="~/.config/chezmoi/chezmoi.toml"
[diff]
    command = "flatpak"
    args = ["run", "com.vscodium.codium", "--wait", "--diff"]
```

Note that the command is `flatpak`, the first two arguments are `run` and the
name of app, and any further arguments are passed to the app.

[atuin]: https://atuin.sh/
[auto-commit]: /user-guide/daily-operations.md#automatically-commit-and-push-changes-to-your-repo
[bash]: https://www.gnu.org/software/bash/
[choose]: https://github.com/twpayne/chezmoi/issues/new/choose
[completion-cmd]: /reference/commands/completion.md
[completion-fun]: /reference/templates/functions/completion.md
[fish]: https://fishshell.com/
[ignore]: /reference/special-files/chezmoiignore.md
[powershell]: https://learn.microsoft.com/en-us/powershell/
[scriptenv]: /user-guide/use-scripts-to-perform-actions.md#set-environment-variables
[zsh]: https://zsh.sourceforge.io/
[flatpak]: https://flatpak.org/
[doctor]: /reference/commands/doctor.md
