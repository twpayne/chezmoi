# A simple assertion library using Go generics

[![PkgGoDev](https://pkg.go.dev/badge/github.com/alecthomas/assert/v2)](https://pkg.go.dev/github.com/alecthomas/assert/v2) [![CI](https://github.com/alecthomas/assert/actions/workflows/ci.yml/badge.svg)](https://github.com/alecthomas/assert/actions/workflows/ci.yml) 
[![Go Report Card](https://goreportcard.com/badge/github.com/alecthomas/assert/v2)](https://goreportcard.com/report/github.com/alecthomas/assert/v2) [![Slack chat](https://img.shields.io/static/v1?logo=slack&style=flat&label=slack&color=green&message=gophers)](https://gophers.slack.com/messages/CN9DS8YF3)


This library is inspired by testify/require, but with a significantly reduced
API surface based on empirical use of that package.

It also provides much nicer diff output, eg.

```
=== RUN   TestFail
    assert_test.go:14: Expected values to be equal:
         assert.Data{
        -  Str: "foo",
        +  Str: "far",
           Num: 10,
         }
--- FAIL: TestFail (0.00s)
```

## API

Import then use as `assert`:

```go
import "github.com/alecthomas/assert/v2"
```

This library has the following API. For all functions, `msgAndArgs` is used to
format error messages using the `fmt` package.

```go
// Equal asserts that "expected" and "actual" are equal using google/go-cmp.
//
// If they are not, a diff of the Go representation of the values will be displayed.
func Equal[T comparable](t testing.TB, expected, actual T, msgAndArgs ...interface{})

// NotEqual asserts that "expected" is not equal to "actual" using google/go-cmp.
//
// If they are equal the expected value will be displayed.
func NotEqual[T comparable](t testing.TB, expected, actual T, msgAndArgs ...interface{})

// Zero asserts that a value is its zero value.
func Zero[T comparable](t testing.TB, value T, msgAndArgs ...interface{})

// NotZero asserts that a value is not its zero value.
func NotZero[T comparable](t testing.TB, value T, msgAndArgs ...interface{})

// Contains asserts that "haystack" contains "needle".
func Contains(t testing.TB, haystack string, needle string, msgAndArgs ...interface{})

// NotContains asserts that "haystack" does not contain "needle".
func NotContains(t testing.TB, haystack string, needle string, msgAndArgs ...interface{})

// EqualError asserts that either an error is non-nil and that its message is what is expected,
// or that error is nil if the expected message is empty.
func EqualError(t testing.TB, err error, errString string, msgAndArgs...interface{})

// Error asserts that an error is not nil.
func Error(t testing.TB, err error, msgAndArgs ...interface{})

// NoError asserts that an error is nil.
func NoError(t testing.TB, err error, msgAndArgs ...interface{})

// IsError asserts than any error in "err"'s tree matches "target".
func IsError(t testing.TB, err, target error, msgAndArgs ...interface{})

// NotIsError asserts than no error in "err"'s tree matches "target".
func NotIsError(t testing.TB, err, target error, msgAndArgs ...interface{})

// Panics asserts that the given function panics.
func Panics(t testing.TB, fn func(), msgAndArgs ...interface{})

// NotPanics asserts that the given function does not panic.
func NotPanics(t testing.TB, fn func(), msgAndArgs ...interface{})

// Compare two values for equality and return true or false.
func Compare[T any](t testing.TB, x, y T) bool

// True asserts that an expression is true.
func True(t testing.TB, ok bool, msgAndArgs ...interface{})

// False asserts that an expression is false.
func False(t testing.TB, ok bool, msgAndArgs ...interface{})
```

## Evaluation process

Our empircal data of testify usage comes from a monorepo with around 50K lines
of tests.

These are the usage counts for all testify functions, normalised to the base
(not `Printf()`) non-negative(not `No(t)?`) case for each core function.

```text
2240 Error
1314 Equal
 219 True
 210 Nil
 167 Empty
 107 Contains
  79 Len
  61 False
  24 EqualValues
  20 EqualError
  17 Zero
  15 Fail
  15 ElementsMatch
   9 Panics
   7 IsType
   6 FileExists
   4 JSONEq
   3 PanicsWithValue
   3 Eventually
```

The decision for each function was:

### Keep

- `Error(t, err)` -> frequently used, keep
- `Equal(t, expected, actual)` -> frequently used, keep but make type safe
- `True(t, expr)` -> frequently used, keep
- `False(t, expr)` -> frequently used, keep
- `Empty(t, thing)` -> `require.Equal(t, len(thing), 0)`
- `Contains(t, haystack string, needle string)` - the only variant used in our codebase, keep as concrete type
- `Zero(t, value)` -> make type safe, keep
- `Panics(t, f)` -> useful, keep
- `EqualError(t, a, b)` -> useful, keep
- `Nil(t, value)` -> frequently used, keep

### Not keeping, replace with ...

- `ElementsMatch(t, a, b)` - use [peterrk/slices](https://github.com/peterrk/slices) or stdlib sort support once it lands.
- `IsType(t, a, b)` -> `require.Equal(t, reflect.TypeOf(a).String(), reflect.TypeOf(b).String())`
- `FileExists()` -> very little use, drop
- `JSONEq()` -> very little use, drop
- `PanicsWithValue()` -> very little use, drop
- `Eventually()` -> very little use, drop
- `Contains(t, haystack []T, needle T)` - very little use, replace with
- `Contains(t, haystack map[K]V, needle K)` - very little use, drop
- `Len(t, v, n)` -> cannot be implemented as a single function with generics`Equal(t, len(v), n)`
- `EqualValues()` - `Equal(t, TYPE(a), TYPE(b))`
- `Fail()` -> `t.Fatal()`
