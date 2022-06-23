# Concepts

chezmoi computes the target state for the current machine and then updates the
destination directory, where:

* The *destination directory* is the directory that chezmoi manages, usually
  your home directory, `~`.

* A *target* is a file, directory, or symlink in the destination directory.

* The *destination state* is the current state of all the targets in the
  destination directory.

* The *source state* declares the desired state of your home directory,
  including templates that use machine-specific data. It contains only regular
  files and directories.

* The *source directory* is where chezmoi stores the source state. By default
  it is `~/.local/share/chezmoi`.

* The *config file* contains machine-specific data. By default it is
  `~/.config/chezmoi/chezmoi.toml`.

* The *target state* is the desired state of the destination directory. It is
  computed from the source state, the config file, and the destination state.
  The target state includes regular files and directories, and may also include
  symbolic links, scripts to be run, and targets to be removed.

* The *working tree* is the git working tree. Normally it is the same as the
  source directory, but can be a parent of the source directory.
