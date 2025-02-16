# Special files

All files in the source state whose name begins with `.` are ignored by default,
unless they are one of the special files listed here. `.chezmoidata.$FORMAT` and
files in `.chezmoidata` folders are read before all other files so that it can
be used in templates.
