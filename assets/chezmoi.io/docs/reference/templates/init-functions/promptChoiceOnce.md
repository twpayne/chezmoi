# `promptChoiceOnce` *map* *path* *prompt* *choices* [*default*]

`promptChoiceOnce` returns the value of *map* at *path* if it exists and is a
string, otherwise it prompts the user for one of *choices* with *prompt* and an
optional *default* using `promptChoice`.
