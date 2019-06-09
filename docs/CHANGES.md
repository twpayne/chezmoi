# chezmoi Changes

* [Upcoming](#upcoming)
  * [`gpgRecipient` config variable changing to `gpg.recipient`](#gpgrecipient-config-variable-changing-to-gpgrecipient)

## Upcoming

### `gpgRecipient` config variable changing to `gpg.recipient`

The `gpgRecipient` config varaible is changing to `gpg.recipient`. To update,
change your config from:

    gpgRecipient = "..."

to:

    [gpg]
      recipient = "..."

Support for the `gpgRecipient` config variable will be removed in version 2.0.0.