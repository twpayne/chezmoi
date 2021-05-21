# chezmoi How-To Guide

<!--- toc --->
* [Perform daily operations](#perform-daily-operations)
  * [Use a hosted repo to manage your dotfiles across multiple machines](#use-a-hosted-repo-to-manage-your-dotfiles-across-multiple-machines)
  * [Pull the latest changes from your repo and apply them](#pull-the-latest-changes-from-your-repo-and-apply-them)
  * [Pull the latest changes from your repo and see what would change, without actually applying the changes](#pull-the-latest-changes-from-your-repo-and-see-what-would-change-without-actually-applying-the-changes)
  * [Automatically commit and push changes to your repo](#automatically-commit-and-push-changes-to-your-repo)
  * [Install chezmoi and your dotfiles on a new machine with a single command](#install-chezmoi-and-your-dotfiles-on-a-new-machine-with-a-single-command)
* [Manage different types of file](#manage-different-types-of-file)
  * [Have chezmoi create a directory, but ignore its contents](#have-chezmoi-create-a-directory-but-ignore-its-contents)
  * [Ensure that a target is removed](#ensure-that-a-target-is-removed)
  * [Manage part, but not all, of a file](#manage-part-but-not-all-of-a-file)
  * [Populate `~/.ssh/authorized_keys` with your public SSH keys from GitHub](#populate-sshauthorized_keys-with-your-public-ssh-keys-from-github)
* [Integrate chezmoi with your editor](#integrate-chezmoi-with-your-editor)
  * [Configure VIM to run `chezmoi apply` whenever you save a dotfile](#configure-vim-to-run-chezmoi-apply-whenever-you-save-a-dotfile)
* [Include dotfiles from elsewhere](#include-dotfiles-from-elsewhere)
  * [Include a subdirectory from another repository, like Oh My Zsh](#include-a-subdirectory-from-another-repository-like-oh-my-zsh)
  * [Handle configuration files which are externally modified](#handle-configuration-files-which-are-externally-modified)
  * [Import archives](#import-archives)
* [Manage machine-to-machine differences](#manage-machine-to-machine-differences)
  * [Use templates](#use-templates)
  * [Ignore files or a directory on different machines](#ignore-files-or-a-directory-on-different-machines)
  * [Use completely different dotfiles on different machines](#use-completely-different-dotfiles-on-different-machines)
  * [Create a config file on a new machine automatically](#create-a-config-file-on-a-new-machine-automatically)
  * [Re-create your config file](#re-create-your-config-file)
  * [Handle different file locations on different systems with the same contents](#handle-different-file-locations-on-different-systems-with-the-same-contents)
  * [Create an archive of your dotfiles](#create-an-archive-of-your-dotfiles)
* [Keep data private](#keep-data-private)
  * [Use 1Password](#use-1password)
  * [Use Bitwarden](#use-bitwarden)
  * [Use gopass](#use-gopass)
  * [Use KeePassXC](#use-keepassxc)
  * [Use Keychain or Windows Credentials Manager](#use-keychain-or-windows-credentials-manager)
  * [Use LastPass](#use-lastpass)
  * [Use pass](#use-pass)
  * [Use Vault](#use-vault)
  * [Use a custom password manager](#use-a-custom-password-manager)
  * [Encrypt whole files](#encrypt-whole-files)
  * [Use a private configuration file and template variables](#use-a-private-configuration-file-and-template-variables)
* [Use scripts to perform actions](#use-scripts-to-perform-actions)
  * [Understand how scripts work](#understand-how-scripts-work)
  * [Install packages with scripts](#install-packages-with-scripts)
  * [Run a script when the contents of another file changes](#run-a-script-when-the-contents-of-another-file-changes)
* [Use chezmoi on macOS](#use-chezmoi-on-macos)
  * [Use `brew bundle` to manage your brews and casks](#use-brew-bundle-to-manage-your-brews-and-casks)
* [Use chezmoi on Windows](#use-chezmoi-on-windows)
  * [Detect Windows Subsystem for Linux (WSL)](#detect-windows-subsystem-for-linux-wsl)
  * [Run a PowerShell script as admin on Windows](#run-a-powershell-script-as-admin-on-windows)
* [Use chezmoi with GitHub Codespaces, Visual Studio Codespaces, or Visual Studio Code Remote - Containers](#use-chezmoi-with-github-codespaces-visual-studio-codespaces-or-visual-studio-code-remote---containers)
* [Use custom tools](#use-custom-tools)
  * [Customize the `diff` command](#customize-the-diff-command)
  * [Use a merge tool other than vimdiff](#use-a-merge-tool-other-than-vimdiff)
* [Migrating to chezmoi from another dotfile manager](#migrating-to-chezmoi-from-another-dotfile-manager)
  * [Migrate from a dotfile manager that uses symlinks](#migrate-from-a-dotfile-manager-that-uses-symlinks)
* [Migrate away from chezmoi](#migrate-away-from-chezmoi)

## Perform daily operations

### Use a hosted repo to manage your dotfiles across multiple machines

chezmoi relies on your version control system and hosted repo to share changes
across multiple machines. You should create a repo on the source code repository
of your choice (e.g. [Bitbucket](https://bitbucket.org),
[GitHub](https://github.com/), or [GitLab](https://gitlab.com), many people call
their repo `dotfiles`) and push the repo in the source directory here. For
example:

    chezmoi cd
    git remote add origin https://github.com/username/dotfiles.git
    git push -u origin main
    exit

On another machine you can checkout this repo:

    chezmoi init https://github.com/username/dotfiles.git

You can then see what would be changed:

    chezmoi diff

If you're happy with the changes then apply them:

    chezmoi apply

The above commands can be combined into a single init, checkout, and apply:

    chezmoi init --apply --verbose https://github.com/username/dotfiles.git

### Pull the latest changes from your repo and apply them

You can pull the changes from your repo and apply them in a single command:

    chezmoi update

This runs `git pull --rebase` in your source directory and then `chezmoi apply`.

### Pull the latest changes from your repo and see what would change, without actually applying the changes

Run:

    chezmoi git pull -- --rebase && chezmoi diff

This runs `git pull --rebase` in your source directory and `chezmoi
diff` then shows the difference between the target state computed from your
source directory and the actual state.

If you're happy with the changes, then you can run

    chezmoi apply

to apply them.

### Automatically commit and push changes to your repo

chezmoi can automatically commit and push changes to your source directory to
your repo. This feature is disabled by default. To enable it, add the following
to your config file:

    [git]
      autoCommit = true
      autoPush = true

Whenever a change is made to your source directory, chezmoi will commit the
changes with an automatically-generated commit message (if `autoCommit` is true)
and push them to your repo (if `autoPush` is true). `autoPush` implies
`autoCommit`, i.e. if `autoPush` is true then chezmoi will auto-commit your
changes. If you only set `autoCommit` to true then changes will be committed but
not pushed.

Be careful when using `autoPush`. If your dotfiles repo is public and you
accidentally add a secret in plain text, that secret will be pushed to your
public repo.

### Install chezmoi and your dotfiles on a new machine with a single command

chezmoi's install script can run `chezmoi init` for you by passing extra
arguments to the newly installed chezmoi binary. If your dotfiles repo is
`github.com/<github-username>/dotfiles` then installing chezmoi, running
`chezmoi init`, and running `chezmoi apply` can be done in a single line of
shell:

    sh -c "$(curl -fsLS git.io/chezmoi)" -- init --apply <github-username>

If your dotfiles repo has a different name to `dotfiles`, or if you host your
dotfiles on a different service, then see the [reference manual for `chezmoi
init`](https://github.com/twpayne/chezmoi/blob/master/docs/REFERENCE.md#init-repo).

For setting up transitory environments (e.g. short-lived Linux containers) you
can install chezmoi, install your dotfiles, and then remove all traces of
chezmoi, including the source directory and chezmoi's configuration directory,
with a single command:

    sh -c "$(curl -fsLS git.io/chezmoi)" -- init --one-shot <github-username>

## Manage different types of file

### Have chezmoi create a directory, but ignore its contents

If you want chezmoi to create a directory, but ignore its contents, say `~/src`,
first run:

    mkdir -p $(chezmoi source-path)/src

This creates the directory in the source state, which means that chezmoi will
create it (if it does not already exist) when you run `chezmoi apply`.

However, as this is an empty directory it will be ignored by git. So, create a
file in the directory in the source state that will be seen by git (so git does
not ignore the directory) but ignored by chezmoi (so chezmoi does not include it
in the target state):

    touch $(chezmoi source-path)/src/.keep

chezmoi automatically creates `.keep` files when you add an empty directory with
`chezmoi add`.

### Ensure that a target is removed

Create a file called `.chezmoiremove` in the source directory containing a list
of patterns of files to remove. When you run

    chezmoi apply --remove

chezmoi will remove anything in the target directory that matches the pattern.
As this command is potentially dangerous, you should run chezmoi in verbose,
dry-run mode beforehand to see what would be removed:

    chezmoi apply --remove --dry-run --verbose

`.chezmoiremove` is interpreted as a template, so you can remove different files
on different machines. Negative matches (patterns prefixed with a `!`) or
targets listed in `.chezmoiignore` will never be removed.

### Manage part, but not all, of a file

chezmoi, by default, manages whole files, but there are two ways to manage just
parts of a file.

Firstly, a `modify_` script receives the current contents of the file on the
standard input and chezmoi reads the target contents of the file from the
script's standard output. This can be used to change parts of a file, for
example using `sed`. Note that if the file does not exist then the standard
input to the `modify_` script will be empty and it is the script's
responsibility to write a complete file to the standard output.

Secondly, if only a small part of the file changes then consider using a
template to re-generate the full contents of the file from the current state.
For example, Kubernetes configurations include a current context that can be
substituted with:

    current-context: {{ output "kubectl" "config" "current-context" | trim }}

### Populate `~/.ssh/authorized_keys` with your public SSH keys from GitHub

chezmoi can retrieve your public SSH keys from GitHub, which can be useful for
populating your `~/.ssh/authorized_keys`. Put the following in your
`~/.local/share/chezmoi/dot_ssh/authorized_keys.tmpl`, where `username` is your
GitHub username:

    {{ range (gitHubKeys "username") -}}
    {{   .Key }}
    {{ end -}}

## Integrate chezmoi with your editor

### Configure VIM to run `chezmoi apply` whenever you save a dotfile

Put the following in your `.vimrc`:

    autocmd BufWritePost ~/.local/share/chezmoi/* ! chezmoi apply --source-path %

## Include dotfiles from elsewhere

### Include a subdirectory from another repository, like Oh My Zsh

To include a subdirectory from another repository, e.g. [Oh My
Zsh](https://github.com/robbyrussell/oh-my-zsh), you cannot use git submodules
because chezmoi uses its own format for the source state and Oh My Zsh is not
distributed in this format. Instead, you can use the `import` command to import
a snapshot from a tarball:

    curl -s -L -o oh-my-zsh-master.tar.gz https://github.com/robbyrussell/oh-my-zsh/archive/master.tar.gz
    chezmoi import --strip-components 1 --destination ${HOME}/.oh-my-zsh oh-my-zsh-master.tar.gz

Add `oh-my-zsh-master.tar.gz` to `.chezmoiignore` if you run these commands in
your source directory so that chezmoi doesn't try to copy the tarball anywhere.

Disable Oh My Zsh auto-updates by setting `DISABLE_AUTO_UPDATE="true"` in
`~/.zshrc`. Auto updates will cause the `~/.oh-my-zsh` directory to drift out of
sync with chezmoi's source state. To update Oh My Zsh, re-run the `curl` and
`chezmoi import` commands above.

### Handle configuration files which are externally modified

Some programs modify their configuration files. When you next run `chezmoi
apply`, any modifications made by the program will be lost.

You can track changes to these files by replacing with a symlink back to a file
in your source directory, which is under version control. Here is a worked
example for VSCode's `settings.json` on Linux:

Copy the configuration file to your source directory:

    cp ~/.config/Code/User/settings.json $(chezmoi source-path)

Tell chezmoi to ignore this file:

    echo settings.json >> $(chezmoi source-path)/.chezmoiignore

Tell chezmoi that `~/.config/Code/User/settings.json` should be a symlink to the
file in your source directory:

    mkdir -p $(chezmoi source-path)/private_dot_config/private_Code/User
    echo -n "{{ .chezmoi.sourceDir }}/settings.json" > $(chezmoi source-path)/private_dot_config/private_Code/User/symlink_settings.json.tmpl

The prefix `private_` is used because the `~/.config` and `~/.config/Code`
directories are private by default.

Apply the changes:

    chezmoi apply -v

Now, when the program modifies its configuration file it will modify the file in
the source state instead.

### Import archives

It is occasionally useful to import entire archives of configuration into your
source state. The `import` command does this. For example, to import the latest
version
[`github.com/robbyrussell/oh-my-zsh`](https://github.com/robbyrussell/oh-my-zsh)
to `~/.oh-my-zsh` run:

    curl -s -L -o oh-my-zsh-master.tar.gz https://github.com/robbyrussell/oh-my-zsh/archive/master.tar.gz
    chezmoi import --strip-components 1 --destination ~/.oh-my-zsh oh-my-zsh-master.tar.gz

Note that this only updates the source state. You will need to run

    chezmoi apply

to update your destination directory.

## Manage machine-to-machine differences

### Use templates

The primary goal of chezmoi is to manage configuration files across multiple
machines, for example your personal macOS laptop, your work Ubuntu desktop, and
your work Linux laptop. You will want to keep much configuration the same across
these, but also need machine-specific configurations for email addresses,
credentials, etc. chezmoi achieves this functionality by using
[`text/template`](https://pkg.go.dev/text/template) for the source state where
needed.

For example, your home `~/.gitconfig` on your personal machine might look like:

    [user]
      email = "john@home.org"

Whereas at work it might be:

    [user]
      email = "john.smith@company.com"

To handle this, on each machine create a configuration file called
`~/.config/chezmoi/chezmoi.toml` defining variables that might vary from machine
to machine. For example, for your home machine:

    [data]
      email = "john@home.org"

Note that all variable names will be converted to lowercase. This is due to a
feature of a library used by chezmoi.

If you intend to store private data (e.g. access tokens) in
`~/.config/chezmoi/chezmoi.toml`, make sure it has permissions `0600`.

If you prefer, you can use any format supported by
[Viper](https://github.com/spf13/viper) for your configuration file. This
includes JSON, YAML, and TOML. Variable names must start with a letter and be
followed by zero or more letters or digits.

Then, add `~/.gitconfig` to chezmoi using the `--autotemplate` flag to turn it
into a template and automatically detect variables from the `data` section
of your `~/.config/chezmoi/chezmoi.toml` file:

    chezmoi add --autotemplate ~/.gitconfig

You can then open the template (which will be saved in the file
`~/.local/share/chezmoi/dot_gitconfig.tmpl`):

    chezmoi edit ~/.gitconfig

The file should look something like:

    [user]
      email = "{{ .email }}"

To disable automatic variable detection, use the `--template` or `-T` option to
`chezmoi add` instead of `--autotemplate`.

Templates are often used to capture machine-specific differences. For example,
in your `~/.local/share/chezmoi/dot_bashrc.tmpl` you might have:

    # common config
    export EDITOR=vi

    # machine-specific configuration
    {{- if eq .chezmoi.hostname "work-laptop" }}
    # this will only be included in ~/.bashrc on work-laptop
    {{- end }}

For a full list of variables, run:

    chezmoi data

For more advanced usage, you can use the full power of the
[`text/template`](https://pkg.go.dev/text/template) language. chezmoi includes
all of the text functions from [sprig](http://masterminds.github.io/sprig/) and
its own [functions for interacting with password
managers](https://github.com/twpayne/chezmoi/blob/master/docs/REFERENCE.md#template-functions).

Templates can be executed directly from the command line, without the need to
create a file on disk, with the `execute-template` command, for example:

    chezmoi execute-template '{{ .chezmoi.os }}/{{ .chezmoi.arch }}'

This is useful when developing or debugging templates.

Some password managers allow you to store complete files. The files can be
retrieved with chezmoi's template functions. For example, if you have a file
stored in 1Password with the UUID `uuid` then you can retrieve it with the
template:

    {{- onepasswordDocument "uuid" -}}

The `-`s inside the brackets remove any whitespace before or after the template
expression, which is useful if your editor has added any newlines.

If, after executing the template, the file contents are empty, the target file
will be removed. This can be used to ensure that files are only present on
certain machines. If you want an empty file to be created anyway, you will need
to give it an `empty_` prefix.

### Ignore files or a directory on different machines

For coarser-grained control of files and entire directories managed on different
machines, or to exclude certain files completely, you can create
`.chezmoiignore` files in the source directory. These specify a list of patterns
that chezmoi should ignore, and are interpreted as templates. An example
`.chezmoiignore` file might look like:

    README.md
    {{- if ne .chezmoi.hostname "work-laptop" }}
    .work # only manage .work on work-laptop
    {{- end }}

The use of `ne` (not equal) is deliberate. What we want to achieve is "only
install `.work` if hostname is `work-laptop`" but chezmoi installs everything by
default, so we have to turn the logic around and instead write "ignore `.work`
unless the hostname is `work-laptop`".

Patterns can be excluded by prefixing them with a `!`, for example:

    f*
    !foo

will ignore all files beginning with an `f` except `foo`.

### Use completely different dotfiles on different machines

chezmoi's template functionality allows you to change a file's contents based on
any variable. For example, if you want `~/.bashrc` to be different on Linux and
macOS you would create a file in the source state called `dot_bashrc.tmpl`
containing:

    {{ if eq .chezmoi.os "darwin" -}}
    # macOS .bashrc contents
    {{ else if eq .chezmoi.os "linux" -}}
    # Linux .bashrc contents
    {{ end -}}

However, if the differences between the two versions are so large that you'd
prefer to use completely separate files in the source state, you can achieve
this using a symbolic link template. Create the following files:

`symlink_dot_bashrc.tmpl`:

    .bashrc_{{ .chezmoi.os }}

`dot_bashrc_darwin`:

    # macOS .bashrc contents

`dot_bashrc_linux`:

    # Linux .bashrc contents

`.chezmoiignore`

    {{ if ne .chezmoi.os "darwin" }}
    .bashrc_darwin
    {{ end }}
    {{ if ne .chezmoi.os "linux" }}
    .bashrc_linux
    {{ end }}

This will make `~/.bashrc` a symlink to `.bashrc_darwin` on `darwin` and to
`.bashrc_linux` on `linux`. The `.chezmoiignore` configuration ensures that only
the OS-specific `.bashrc_os` file will be installed on each OS.

#### Without using symlinks

The same thing can be achieved using the include function.

`dot_bashrc.tmpl`

	{{ if eq .chezmoi.os "darwin" }}
	{{   include ".bashrc_darwin" }}
	{{ end }}
	{{ if eq .chezmoi.os "linux" }}
	{{   include ".bashrc_linux" }}
	{{ end }}

### Create a config file on a new machine automatically

`chezmoi init` can also create a config file automatically, if one does not
already exist. If your repo contains a file called `.chezmoi.<format>.tmpl`
where *format* is one of the supported config file formats (e.g. `json`, `toml`,
or `yaml`) then `chezmoi init` will execute that template to generate your
initial config file.

Specifically, if you have `.chezmoi.toml.tmpl` that looks like this:

    {{- $email := promptString "email" -}}
    [data]
        email = {{ $email | quote }}

Then `chezmoi init` will create an initial `chezmoi.toml` using this template.
`promptString` is a special function that prompts the user (you) for a value.

To test this template, use `chezmoi execute-template` with the `--init` and
`--promptString` flags, for example:

    chezmoi execute-template --init --promptString email=john@home.org < ~/.local/share/chezmoi/.chezmoi.toml.tmpl

### Re-create your config file

If you change your config file template, chezmoi will warn you if your current
config file was not generated from that template. You can re-generate your
config file by running:

    chezmoi init

If you are using any `prompt*` template functions in your config file template
you will be prompted again. However, you can avoid this with the following
example template logic:

    {{- $email := get . "email" -}}
    {{- if not $email -}}
    {{-   $email = promptString "email" -}}
    {{- end -}}

    [data]
      email = {{ $email | quote }}

This will cause chezmoi to first try to re-use the existing `$email` variable
and fallback to `promptString` only if it is not set.

For boolean variables you can use:

    {{- $var := false -}}
    {{- if (hasKey . "var") -}}
    {{-   $var = get . "var" -}}
    {{- end -}}

    [data]
      var = $var

### Handle different file locations on different systems with the same contents

If you want to have the same file contents in different locations on different
systems, but maintain only a single file in your source state, you can use
a shared template.

Create the common file in the `.chezmoitemplates` directory in the source state. For
example, create `.chezmoitemplates/file.conf`. The contents of this file are
available in templates with the `template *name*` function where *name* is the
name of the file.

Then create files for each system, for example `Library/Application
Support/App/file.conf.tmpl` for macOS and `dot_config/app/file.conf.tmpl` for
Linux. Both template files should contain `{{- template "file.conf" -}}`.

Finally, tell chezmoi to ignore files where they are not needed by adding lines
to your `.chezmoiignore` file, for example:

    {{ if ne .chezmoi.os "darwin" }}
    Library/Application Support/App/file.conf
    {{ end }}
    {{ if ne .chezmoi.os "linux" }}
    .config/app/file.conf
    {{ end }}

### Create an archive of your dotfiles

`chezmoi archive` creates an archive containing the target state. This can be
useful for generating target state for a different machine. You can specify a
different configuration file (including template variables) with the `--config`
option.

## Keep data private

chezmoi automatically detects when files and directories are private when adding
them by inspecting their permissions. Private files and directories are stored
in `~/.local/share/chezmoi` as regular, public files with permissions `0644` and
the name prefix `private_`. For example:

    chezmoi add ~/.netrc

will create `~/.local/share/chezmoi/private_dot_netrc` (assuming `~/.netrc` is
not world- or group- readable, as it should be). This file is still private
because `~/.local/share/chezmoi` is not group- or world- readable or executable.
chezmoi checks that the permissions of `~/.local/share/chezmoi` are `0700` on
every run and will print a warning if they are not.

It is common that you need to store access tokens in config files, e.g. a
[GitHub access
token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).
There are several ways to keep these tokens secure, and to prevent them leaving
your machine.

### Use 1Password

chezmoi includes support for [1Password](https://1password.com/) using the
[1Password CLI](https://support.1password.com/command-line-getting-started/) to
expose data as a template function.

Log in and get a session using:

    eval $(op signin <subdomain>.1password.com <email>)

The output of `op get item <uuid>` is available as the `onepassword` template
function. chezmoi parses the JSON output and returns it as structured data. For
example, if the output of `op get item "<uuid>"` is:

    {
        "uuid": "<uuid>",
        "details": {
            "password": "xxx"
        }
    }

Then you can access `details.password` with the syntax:

    {{ (onepassword "<uuid>").details.password }}

Login details fields can be retrieved with the `onepasswordDetailsFields`
function, for example:

    {{- (onepasswordDetailsFields "uuid").password.value }}

Documents can be retrieved with:

    {{- onepasswordDocument "uuid" -}}

Note the extra `-` after the opening `{{` and before the closing `}}`. This
instructs the template language to remove and whitespace before and after the
substitution. This removes any trailing newline added by your editor when saving
the template.

### Use Bitwarden

chezmoi includes support for [Bitwarden](https://bitwarden.com/) using the
[Bitwarden CLI](https://github.com/bitwarden/cli) to expose data as a template
function.

Log in to Bitwarden using:

    bw login <bitwarden-email>

Unlock your Bitwarden vault:

    bw unlock

Set the `BW_SESSION` environment variable, as instructed.

The structured data from `bw get` is available as the `bitwarden` template
function in your config files, for example:

    username = {{ (bitwarden "item" "example.com").login.username }}
    password = {{ (bitwarden "item" "example.com").login.password }}

Custom fields can be accessed with the `bitwardenFields` template function. For
example, if you have a custom field named `token` you can retrieve its value
with:

    {{ (bitwardenFields "item" "example.com").token.value }}

### Use gopass

chezmoi includes support for [gopass](https://www.gopass.pw/) using the gopass CLI.

The first line of the output of `gopass show <pass-name>` is available as the
`gopass` template function, for example:

    {{ gopass "<pass-name>" }}

### Use KeePassXC

chezmoi includes support for [KeePassXC](https://keepassxc.org) using the
KeePassXC CLI (`keepassxc-cli`) to expose data as a template function.

Provide the path to your KeePassXC database in your configuration file:

    [keepassxc]
      database = "/home/user/Passwords.kdbx"

The structured data from `keepassxc-cli show $database` is available as the
`keepassxc` template function in your config files, for example:

    username = {{ (keepassxc "example.com").UserName }}
    password = {{ (keepassxc "example.com").Password }}

Additional attributes are available through the `keepassxcAttribute` function.
For example, if you have an entry called `SSH Key` with an additional attribute
called `private-key`, its value is available as:

    {{ keepassxcAttribute "SSH Key" "private-key" }}

### Use Keychain or Windows Credentials Manager

chezmoi includes support for Keychain (on macOS), GNOME Keyring (on Linux), and
Windows Credentials Manager (on Windows) via the
[`zalando/go-keyring`](https://github.com/zalando/go-keyring) library.

Set values with:

    $ chezmoi secret keyring set --service=<service> --user=<user>
    Value: xxxxxxxx

The value can then be used in templates using the `keyring` function which takes
the service and user as arguments.

For example, save a GitHub access token in keyring with:

    $ chezmoi secret keyring set --service=github --user=<github-username>
    Value: xxxxxxxx

and then include it in your `~/.gitconfig` file with:

    [github]
      user = "{{ .github.user }}"
      token = "{{ keyring "github" .github.user }}"

You can query the keyring from the command line:

    chezmoi secret keyring get --service=github --user=<github-username>

### Use LastPass

chezmoi includes support for [LastPass](https://lastpass.com) using the
[LastPass CLI](https://lastpass.github.io/lastpass-cli/lpass.1.html) to expose
data as a template function.

Log in to LastPass using:

    lpass login <lastpass-username>

Check that `lpass` is working correctly by showing password data:

    lpass show --json <lastpass-entry-id>

where `<lastpass-entry-id>` is a [LastPass Entry
Specification](https://lastpass.github.io/lastpass-cli/lpass.1.html#_entry_specification).

The structured data from `lpass show --json id` is available as the `lastpass`
template function. The value will be an array of objects. You can use the
`index` function and `.Field` syntax of the `text/template` language to extract
the field you want. For example, to extract the `password` field from first the
"GitHub" entry, use:

    githubPassword = "{{ (index (lastpass "GitHub") 0).password }}"

chezmoi automatically parses the `note` value of the Lastpass entry as
colon-separated key-value pairs, so, for example, you can extract a private SSH
key like this:

    {{ (index (lastpass "SSH") 0).note.privateKey }}

Keys in the `note` section written as `CamelCase Words` are converted to
`camelCaseWords`.

If the `note` value does not contain colon-separated key-value pairs, then you
can use `lastpassRaw` to get its raw value, for example:

    {{ (index (lastpassRaw "SSH Private Key") 0).note }}

### Use pass

chezmoi includes support for [pass](https://www.passwordstore.org/) using the
pass CLI.

The first line of the output of `pass show <pass-name>` is available as the
`pass` template function, for example:

    {{ pass "<pass-name>" }}

### Use Vault

chezmoi includes support for [Vault](https://www.vaultproject.io/) using the
[Vault CLI](https://www.vaultproject.io/docs/commands/) to expose data as a
template function.

The vault CLI needs to be correctly configured on your machine, e.g. the
`VAULT_ADDR` and `VAULT_TOKEN` environment variables must be set correctly.
Verify that this is the case by running:

    vault kv get -format=json <key>

The structured data from `vault kv get -format=json` is available as the `vault`
template function. You can use the `.Field` syntax of the `text/template`
language to extract the data you want. For example:

    {{ (vault "<key>").data.data.password }}

### Use a custom password manager

You can use any command line tool that outputs secrets either as a string or in
JSON format. Choose the binary by setting `secret.command` in your configuration
file. You can then invoke this command with the `secret` and `secretJSON`
template functions which return the raw output and JSON-decoded output
respectively. All of the above secret managers can be supported in this way:

| Secret Manager  | `secret.command` | Template skeleton                                 |
| --------------- | ---------------- | ------------------------------------------------- |
| 1Password       | `op`             | `{{ secretJSON "get" "item" <id> }}`              |
| Bitwarden       | `bw`             | `{{ secretJSON "get" <id> }}`                     |
| HashiCorp Vault | `vault`          | `{{ secretJSON "kv" "get" "-format=json" <id> }}` |
| LastPass        | `lpass`          | `{{ secretJSON "show" "--json" <id> }}`           |
| KeePassXC       | `keepassxc-cli`  | Not possible (interactive command only)           |
| pass            | `pass`           | `{{ secret "show" <id> }}`                        |

### Encrypt whole files

#### gpg

chezmoi supports encrypting files with [gpg](https://www.gnupg.org/). Encrypted
files are stored in the source state and automatically be decrypted when
generating the target state or printing a file's contents with `chezmoi cat`.
`chezmoi edit` will transparently decrypt the file before editing and re-encrypt
it afterwards.

##### Asymmetric (private/public-key) encryption

Specify the encryption key to use in your configuration file (`chezmoi.toml`)
with the `gpg.recipient` key:

    encryption = "gpg"
    [gpg]
      recipient = "..."

Add files to be encrypted with the `--encrypt` flag, for example:

    chezmoi add --encrypt ~/.ssh/id_rsa

chezmoi will encrypt the file with:

    gpg --armor --recipient ${gpg.recipient} --encrypt

and store the encrypted file in the source state. The file will automatically be
decrypted when generating the target state.

##### Symmetric encryption

Specify symmetric encryption in your configuration file:

    encryption = "gpg"
    [gpg]
      symmetric = true

Add files to be encrypted with the `--encrypt` flag, for example:

    chezmoi add --encrypt ~/.ssh/id_rsa

chezmoi will encrypt the file with:

    gpg --armor --symmetric

### Use a private configuration file and template variables

Typically, `~/.config/chezmoi/chezmoi.toml` is not checked in to version control
and has permissions 0600. You can store tokens as template values in the `data`
section. For example, if your `~/.config/chezmoi/chezmoi.toml` contains:

    [data]
      [data.github]
        user = "<github-username>"
        token = "<github-token>"

Your `~/.local/share/chezmoi/private_dot_gitconfig.tmpl` can then contain:

    {{- if (index . "github") }}
    [github]
      user = "{{ .github.user }}"
      token = "{{ .github.token }}"
    {{- end }}

Any config files containing tokens in plain text should be private (permissions
`0600`).

## Use scripts to perform actions

### Understand how scripts work

chezmoi supports scripts, which are executed when you run `chezmoi apply`. The
scripts can either run every time you run `chezmoi apply`, or only when their
contents have changed.

In verbose mode, the script's contents will be printed before executing it. In
dry-run mode, the script is not executed.

Scripts are any file in the source directory with the prefix `run_`, and are
executed in alphabetical order. Scripts that should only be run when their
contents change have the prefix `run_once_`.

Scripts break chezmoi's declarative approach, and as such should be used
sparingly. Any script should be idempotent, even `run_once_` scripts.

Scripts must be created manually in the source directory, typically by running
`chezmoi cd` and then creating a file with a `run_` prefix. Scripts are executed
directly using `exec` and must include a shebang line or be executable binaries.
There is no need to set the executable bit on the script.

Scripts with the suffix `.tmpl` are treated as templates, with the usual
template variables available. If, after executing the template, the result is
only whitespace or an empty string, then the script is not executed. This is
useful for disabling scripts.

### Install packages with scripts

Change to the source directory and create a file called
`run_once_install-packages.sh`:

    chezmoi cd
    $EDITOR run_once_install-packages.sh

In this file create your package installation script, e.g.

    #!/bin/sh
    sudo apt install ripgrep

The next time you run `chezmoi apply` or `chezmoi update` this script will be
run. As it has the `run_once_` prefix, it will not be run again unless its
contents change, for example if you add more packages to be installed.

This script can also be a template. For example, if you create
`run_once_install-packages.sh.tmpl` with the contents:

    {{ if eq .chezmoi.os "linux" -}}
    #!/bin/sh
    sudo apt install ripgrep
    {{ else if eq .chezmoi.os "darwin" -}}
    #!/bin/sh
    brew install ripgrep
    {{ end -}}

This will install `ripgrep` on both Debian/Ubuntu Linux systems and macOS.

### Run a script when the contents of another file changes

chezmoi's `run_` scripts are run every time you run `chezmoi apply`, whereas
`run_once_` scripts are run only when their contents have changed, after
executing them as templates. You use this to cause a `run_once_` script to run
when the contents of another file has changed by including a checksum of the
other file's contents in the script.

For example, if your [dconf](https://wiki.gnome.org/Projects/dconf) settings are
stored in `dconf.ini` in your source directory then you can make `chezmoi apply`
only load them when the contents of `dconf.ini` has changed by adding the
following script as `run_once_dconf-load.sh.tmpl`:

    #!/bin/bash

    # dconf.ini hash: {{ include "dconf.ini" | sha256sum }}
    dconf load / "{{ joinPath .chezmoi.sourceDir "dconf.ini" }}"

As the SHA256 sum of `dconf.ini` is included in a comment in the script, the
contents of the script will change whenever the contents of `dconf.ini` are
changed, so chezmoi will re-run the script whenever the contents of `dconf.ini`
change.

In this example you should also add `dconf.ini` to `.chezmoiignore` so chezmoi
does not create `dconf.ini` in your home directory.

## Use chezmoi on macOS

### Use `brew bundle` to manage your brews and casks

Homebrew's [`brew bundle`
subcommand](https://docs.brew.sh/Manpage#bundle-subcommand) allows you to
specify a list of brews and casks to be installed. You can integrate this with
chezmoi by creating a `run_once_` script. For example, create a file in your
source directory called `run_once_before_install-packages-darwin.sh.tmpl`
containing:

    {{- if (eq .chezmoi.os "darwin") -}}
    #!/bin/bash

    brew bundle --no-lock --file=/dev/stdin <<EOF
    brew "git"
    cask "google-chrome"
    EOF
    {{ end -}}

Note that the `Brewfile` is embedded directly in the script with a bash here
document. chezmoi will run this script whenever its contents change, i.e. when
you add or remove brews or casks.

## Use chezmoi on Windows

### Detect Windows Subsystem for Linux (WSL)

WSL can be detected by looking for the string `Microsoft` or `microsoft` in
`/proc/sys/kernel/osrelease`, which is available in the template variable
`.chezmoi.kernel.osrelease`, for example:

    {{ if (eq .chezmoi.os "linux") }}
    {{   if (.chezmoi.kernel.osrelease | lower | contains "microsoft") }}
    # WSL-specific code
    {{   end }}
    {{ end }}

### Run a PowerShell script as admin on Windows

Put the following at the top of your script:

```powershell
# Self-elevate the script if required
if (-Not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] 'Administrator')) {
  if ([int](Get-CimInstance -Class Win32_OperatingSystem | Select-Object -ExpandProperty BuildNumber) -ge 6000) {
    $CommandLine = "-NoExit -File `"" + $MyInvocation.MyCommand.Path + "`" " + $MyInvocation.UnboundArguments
    Start-Process -FilePath PowerShell.exe -Verb Runas -ArgumentList $CommandLine
    Exit
  }
}
```

## Use chezmoi with GitHub Codespaces, Visual Studio Codespaces, or Visual Studio Code Remote - Containers

The following assumes you are using chezmoi 1.8.4 or later. It does not work
with earlier versions of chezmoi.

You can use chezmoi to manage your dotfiles in [GitHub
Codespaces](https://docs.github.com/en/github/developing-online-with-codespaces/personalizing-codespaces-for-your-account),
[Visual Studio
Codespaces](https://docs.microsoft.com/en/visualstudio/codespaces/reference/personalizing),
and [Visual Studio Code Remote -
Containers](https://code.visualstudio.com/docs/remote/containers#_personalizing-with-dotfile-repositories).

For a quick start, you can clone the [`chezmoi/dotfiles`
repository](https://github.com/chezmoi/dotfiles) which supports Codespaces out
of the box.

The workflow is different to using chezmoi on a new machine, notably:
* These systems will automatically clone your `dotfiles` repo to `~/dotfiles`,
  so there is no need to clone your repo yourself.
* The installation script must be non-interactive.
* When running in a Codespace, the environment variable `CODESPACES` will be set
  to `true`. You can read its value with the [`env` template
  function](http://masterminds.github.io/sprig/os.html).

First, if you are using a chezmoi configuration file template, ensure that it is
non-interactive when running in Codespaces, for example, `.chezmoi.toml.tmpl`
might contain:

    {{- $codespaces:= env "CODESPACES" | not | not -}}
    sourceDir = "{{ .chezmoi.sourceDir }}"

    [data]
      name = "Your name"
      codespaces = {{ $codespaces }}
    {{- if $codespaces }}{{/* Codespaces dotfiles setup is non-interactive, so set an email address */}}
      email = "your@email.com"
    {{- else }}{{/* Interactive setup, so prompt for an email address */}}
      email = "{{ promptString "email" }}"
    {{- end }}

This sets the `codespaces` template variable, so you don't have to repeat `(env
"CODESPACES")` in your templates. It also sets the `sourceDir` configuration to
the `--source` argument passed in `chezmoi init`.

Second, create an `install.sh` script that installs chezmoi and your dotfiles:

```sh
#!/bin/sh

set -e # -e: exit on error

if [ ! "$(command -v chezmoi)" ]; then
  bin_dir="$HOME/.local/bin"
  chezmoi="$bin_dir/chezmoi"
  if [ "$(command -v curl)" ]; then
    sh -c "$(curl -fsLS https://git.io/chezmoi)" -- -b "$bin_dir"
  elif [ "$(command -v wget)" ]; then
    sh -c "$(wget -qO- https://git.io/chezmoi)" -- -b "$bin_dir"
  else
    echo "To install chezmoi, you must have curl or wget installed." >&2
    exit 1
  fi
else
  chezmoi=chezmoi
fi

# POSIX way to get script's dir: https://stackoverflow.com/a/29834779/12156188
script_dir="$(cd -P -- "$(dirname -- "$(command -v -- "$0")")" && pwd -P)"
# exec: replace current process with chezmoi init
exec "$chezmoi" init --apply "--source=$script_dir"
```

Ensure that this file is executable (`chmod a+x install.sh`), and add
`install.sh` to your `.chezmoiignore` file.

It installs the latest version of chezmoi in `~/.local/bin` if needed, and then
`chezmoi init ...` invokes chezmoi to create its configuration file and
initialize your dotfiles. `--apply` tells chezmoi to apply the changes
immediately, and `--source=...` tells chezmoi where to find the cloned
`dotfiles` repo, which in this case is the same folder in which the script is
running from.

If you do not use a chezmoi configuration file template you can use `chezmoi
apply --source=$HOME/dotfiles` instead of `chezmoi init ...` in `install.sh`.

Finally, modify any of your templates to use the `codespaces` variable if
needed. For example, to install `vim-gtk` on Linux but not in Codespaces, your
`run_once_install-packages.sh.tmpl` might contain:

    {{- if (and (eq .chezmoi.os "linux") (not .codespaces)) -}}
    #!/bin/sh
    sudo apt install -y vim-gtk
    {{- end -}}

## Use custom tools

### Customize the `diff` command

By default, chezmoi uses a built-in diff. You can change the format, and/or pipe
the output into a pager of your choice. For example, to use
[`diff-so-fancy`](https://github.com/so-fancy/diff-so-fancy) specify:

    [diff]
      pager = "diff-so-fancy"

The pager can be disabled using the `--no-pager` flag.

### Use a merge tool other than vimdiff

By default, chezmoi uses vimdiff, but you can use any merge tool of your choice.
In your config file, specify the command and args to use. For example, to use
neovim's diff mode specify:

    [merge]
      command = "nvim"
      args = "-d"

## Migrating to chezmoi from another dotfile manager

### Migrate from a dotfile manager that uses symlinks

Many dotfile managers replace dotfiles with symbolic links to files in a common
directory. If you `chezmoi add` such a symlink, chezmoi will add the symlink,
not the file. To assist with migrating from symlink-based systems, use the
`--follow` option to `chezmoi add`, for example:

    chezmoi add --follow ~/.bashrc

This will tell `chezmoi add` that the target state of `~/.bashrc` is the target
of the `~/.bashrc` symlink, rather than the symlink itself. When you run
`chezmoi apply`, chezmoi will replace the `~/.bashrc` symlink with the file
contents.

## Migrate away from chezmoi

chezmoi creates regular files and directories - your home directory managed by
chezmoi is exactly the same as if you were using no dotfile manager at all. Use
your new dotfile manager's functionality for importing your existing dotfiles.
