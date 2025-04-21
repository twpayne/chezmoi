# `toString` *value*

`toString` returns the string representation of *value*. Notably, if *value* is
a pointer, then it is safely dereferenced. If *value* is a nil pointer, then
`toString` returns the string representation of the zero value of the pointee's
type.
