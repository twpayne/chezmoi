# `fromJson` *jsontext*

`fromJson` parses *jsontext* as JSON and returns the parsed value.

JSON numbers that can be represented exactly as 64-bit signed integers are
returned as such. Otherwise, if the number is in the range of 64-bit IEEE
floating point values, it is returned as such. Otherwise, the number is returned
as a string. See [RFC7159 Section 6][rfc7159s6].

[rfc7159S6]: https://www.rfc-editor.org/rfc/rfc7159#section-6
