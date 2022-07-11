# passhole function

The `passhole` template function returns structured data from a [Keepass](https://keepass.info/) database using the [Passhole Cli](https://github.com/Evidlo/passhole) (`ph`).

The database path is configured by setting the `passhole.database` configuration variable.
There are other configuration variables:
  - keyfile: If you created a database with a keyfile, this is required to open it.
  - nocache: Boolean, whether passhole cache your password or not.
  - nopassword: Boolean
