# Setup

## Understand chezmoi's files and directories

chezmoi generates your dotfiles for your local machine. It combines two main
sources of data:

The *source directory*, `~/.local/share/chezmoi`, is common to all your
machines, and is a clone of your dotfiles repo. Each file that chezmoi manages
has a corresponding file in the source directory.

The *config file*, typically `~/.config/chezmoi/chezmoi.toml` (although you can
use JSON or YAML if you prefer), is specific to the local machine.

Files whose contents are the same on all of your machines are copied verbatim
from the source directory. Files which vary from machine to machine are executed
as templates, typically using data from the local machine's config file to tune
the final contents specific to the local machine.

## Use a hosted repo to manage your dotfiles across multiple machines

chezmoi relies on your version control system and hosted repo to share changes
across multiple machines. You should create a repo on the source code repository
of your choice (e.g. [Bitbucket](https://bitbucket.org),
[GitHub](https://github.com/), or [GitLab](https://gitlab.com), many people call
their repo `dotfiles`) and push the repo in the source directory here. For
example:

```sh
chezmoi cd
git remote add origin https://github.com/$GITHUB_USERNAME/dotfiles.git
git push -u origin main
exit
```

On another machine you can checkout this repo:

```sh
chezmoi init https://github.com/$GITHUB_USERNAME/dotfiles.git
```

You can then see what would be changed:

```sh
chezmoi diff
```

If you're happy with the changes then apply them:

```sh
chezmoi apply
```

The above commands can be combined into a single init, checkout, and apply:

```sh
chezmoi init --apply --verbose https://github.com/$GITHUB_USERNAME/dotfiles.git
```

These commands are summarized this sequence diagram:

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    R->>W: chezmoi init &lt;repo&gt;
    W-->>H: chezmoi diff
    W->>H: chezmoi apply
    R->>H: chezmoi init --apply &lt;repo&gt;
```

## Use a private repo to store your dotfiles

chezmoi supports storing your dotfiles in both public and private repos.

chezmoi is designed so that your dotfiles repo can be public by making it easy
for you to store your secrets either in your password manager, in encrypted
files, or in private configuration files. Your dotfiles repo can still be
private, if you choose.

If you use a private repo for your dotfiles then you will typically need to
enter your credentials (e.g. your username and password) each time you interact
with the repo, for example when pulling or pushing changes. chezmoi itself does
not store any credentials, but instead relies on your local git configuration
for these operations.

When using a private repo on GitHub without `--ssh`, when prompted for a
password you will need to enter a [GitHub personal access
token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token).
For more information on these changes, read the [GitHub blog post on Token
authentication requirements for Git
operations](https://github.blog/2020-12-15-token-authentication-requirements-for-git-operations/)

## Create a config file on a new machine automatically

`chezmoi init` can also create a config file automatically, if one does not
already exist. If your repo contains a file called `.chezmoi.$FORMAT.tmpl`
where `$FORMAT` is one of the supported config file formats (e.g. `json`,
`jsonc`, `toml`, or `yaml`) then `chezmoi init` will execute that template to
generate your initial config file.

Specifically, if you have `.chezmoi.toml.tmpl` that looks like this:

``` title="~/.local/share/chezmoi/.chezmoi.toml.tmpl"
{{- $email := promptStringOnce . "email" "Email address" -}}

[data]
    email = {{ $email | quote }}
```

Then `chezmoi init` will create an initial `chezmoi.toml` using this template.
`promptStringOnce` is a special function that prompts the user (you) for a value
if it is not already set in your `data`.

To test this template, use `chezmoi execute-template` with the `--init` and
`--promptString` flags, for example:

```sh
chezmoi execute-template --init --promptString email=me@home.org < ~/.local/share/chezmoi/.chezmoi.toml.tmpl
```

## Re-create your config file

If you change your config file template, chezmoi will warn you if your current
config file was not generated from that template. You can re-generate your
config file by running:

```sh
chezmoi init
```

If you are using any `prompt*` template functions in your config file template
you will be prompted again. However, you can avoid this with the following
example template logic:

```text
{{- $email := promptStringOnce . "email" "Email address" -}}

[data]
    email = {{ $email | quote }}
```

This will cause chezmoi use the `email` variable from your `data` and fallback
to `promptString` only if it is not set.
