# combinator

[![Latest Release](https://img.shields.io/github/release/muesli/combinator.svg)](https://github.com/muesli/combinator/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/muesli/combinator)
[![Build Status](https://github.com/muesli/combinator/workflows/build/badge.svg)](https://github.com/muesli/combinator/actions)
[![Coverage Status](https://coveralls.io/repos/github/muesli/combinator/badge.svg?branch=master)](https://coveralls.io/github/muesli/combinator?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/muesli/combinator)](http://goreportcard.com/report/muesli/combinator)

`combinator` generates a slice of all possible value combinations for any given
struct and a set of its potential member values. This can be used to generate
extensive test matrixes among other things.

## Installation

```bash
go get github.com/muesli/combinator
```

## Example

```go
type User struct {
    Name  string
    Age   uint
    Admin bool
}

/*
  Define potential test values. Make sure the struct's fields share the name and
  type of the structs you want to generate.
*/
testData := struct {
    Name  []string
    Age   []uint
    Admin []bool
}{
    Name:  []string{"Alice", "Bob"},
    Age:   []uint{23, 42, 99},
    Admin: []bool{false, true},
}

// Generate all possible combinations
var users []User
combinator.Generate(&users, testData)

for i, u := range users {
    fmt.Printf("Combination %2d | Name: %-5s | Age: %d | Admin: %v\n", i, u.Name, u.Age, u.Admin)
}
```

```
Combination  0 | Name: Alice | Age: 23 | Admin: false
Combination  1 | Name: Bob   | Age: 23 | Admin: false
Combination  2 | Name: Alice | Age: 42 | Admin: false
Combination  3 | Name: Bob   | Age: 42 | Admin: false
Combination  4 | Name: Alice | Age: 99 | Admin: false
Combination  5 | Name: Bob   | Age: 99 | Admin: false
Combination  6 | Name: Alice | Age: 23 | Admin: true
Combination  7 | Name: Bob   | Age: 23 | Admin: true
Combination  8 | Name: Alice | Age: 42 | Admin: true
Combination  9 | Name: Bob   | Age: 42 | Admin: true
Combination 10 | Name: Alice | Age: 99 | Admin: true
Combination 11 | Name: Bob   | Age: 99 | Admin: true
```

## License

[MIT](https://github.com/muesli/combinator/raw/master/LICENSE)
