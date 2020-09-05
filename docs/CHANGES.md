# chezmoi Changes

<!--- toc --->
* [Upcoming](#upcoming)
  * [Default diff format changing from `chezmoi` to `git`.](#default-diff-format-changing-from-chezmoi-to-git)
  * [`gpgRecipient` config variable changing to `gpg.recipient`](#gpgrecipient-config-variable-changing-to-gpgrecipient)

## Upcoming

### Default diff format changing from `chezmoi` to `git`.

Currently chezmoi outputs diffs in its own format, containing a mix of unified
diffs and shell commands. This will be replaced with a [git format
diff](https://git-scm.com/docs/diff-format) in version 2.0.0.

### `gpgRecipient` config variable changing to `gpg.recipient`

The `gpgRecipient` config variable is changing to `gpg.recipient`. To update,
change your config from:

    gpgRecipient = "..."

to:

    [gpg]
      recipient = "..."

Support for the `gpgRecipient` config variable will be removed in version 2.0.0.
