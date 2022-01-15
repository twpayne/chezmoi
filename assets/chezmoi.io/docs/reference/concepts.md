# Concepts

chezmoi evaluates the source state for the current machine and then updates the
destination directory, where:

* The *source state* declares the desired state of your home directory,
  including templates and machine-specific configuration.

* The *source directory* is where chezmoi stores the source state, by default
  `~/.local/share/chezmoi`.

* The *target state* is the source state computed for the current machine.

* The *destination directory* is the directory that chezmoi manages, by default
  your home directory.

* A *target* is a file, directory, or symlink in the destination directory.

* The *destination state* is the current state of all the targets in the
  destination directory.

* The *config file* contains machine-specific configuration, by default it is
  `~/.config/chezmoi/chezmoi.toml`.

* The *working tree* is the git working tree. Normally it is the same as the
  source directory, but can be a parent of the source directory.
