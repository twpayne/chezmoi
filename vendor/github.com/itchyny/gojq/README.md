# gojq
[![CI Status](https://github.com/itchyny/gojq/workflows/CI/badge.svg)](https://github.com/itchyny/gojq/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/itchyny/gojq)](https://goreportcard.com/report/github.com/itchyny/gojq)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/itchyny/gojq/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/itchyny/gojq/all.svg)](https://github.com/itchyny/gojq/releases)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/itchyny/gojq)](https://pkg.go.dev/github.com/itchyny/gojq)

### Pure Go implementation of [jq](https://github.com/jqlang/jq)
This is an implementation of jq command written in Go language.
You can also embed gojq as a library to your Go products.

## Usage
```sh
 $ echo '{"foo": 128}' | gojq '.foo'
128
 $ echo '{"a": {"b": 42}}' | gojq '.a.b'
42
 $ echo '{"id": "sample", "10": {"b": 42}}' | gojq '{(.id): .["10"].b}'
{
  "sample": 42
}
 $ echo '[{"id":1},{"id":2},{"id":3}]' | gojq '.[] | .id'
1
2
3
 $ echo '{"a":1,"b":2}' | gojq '.a += 1 | .b *= 2'
{
  "a": 2,
  "b": 4
}
 $ echo '{"a":1} [2] 3' | gojq '. as {$a} ?// [$a] ?// $a | $a'
1
2
3
 $ echo '{"foo": 4722366482869645213696}' | gojq .foo
4722366482869645213696  # keeps the precision of large numbers
 $ gojq -n 'def fact($n): if $n < 1 then 1 else $n * fact($n - 1) end; fact(50)'
30414093201713378043612608166064768844377641568960512000000000000 # arbitrary-precision integer calculation
```

Nice error messages.
```sh
 $ echo '[1,2,3]' | gojq '.foo & .bar'
gojq: invalid query: .foo & .bar
    .foo & .bar
         ^  unexpected token "&"
 $ echo '{"foo": { bar: [] } }' | gojq '.'
gojq: invalid json: <stdin>
    {"foo": { bar: [] } }
              ^  invalid character 'b' looking for beginning of object key string
```

## Installation
### Homebrew
```sh
brew install gojq
```

### Zero Install
```sh
0install add gojq https://apps.0install.net/utils/gojq.xml
```

### Build from source
```sh
go install github.com/itchyny/gojq/cmd/gojq@latest
```

### Docker
```sh
docker run -i --rm itchyny/gojq
docker run -i --rm ghcr.io/itchyny/gojq
```

## Difference to jq
- gojq is purely implemented with Go language and is completely portable. jq depends on the C standard library so the availability of math functions depends on the library. jq also depends on the regular expression library and it makes build scripts complex.
- gojq implements nice error messages for invalid query and JSON input. The error message of jq is sometimes difficult to tell where to fix the query.
- gojq does not keep the order of object keys. I understand this might cause problems for some scripts but basically, we should not rely on the order of object keys. Due to this limitation, gojq does not have `keys_unsorted` function and `--sort-keys` (`-S`) option. I would implement when ordered map is implemented in the standard library of Go but I'm less motivated.
- gojq supports arbitrary-precision integer calculation while jq does not; jq loses the precision of large integers when calculation is involved. Note that even with gojq, all mathematical functions, including `floor` and `round`, convert integers to floating-point numbers; only addition, subtraction, multiplication, modulo, and division operators (when divisible) keep the integer precision. To calculate floor division of integers without losing the precision, use `def idivide($n): (. - . % $n) / $n;`. To round down floating-point numbers to integers, use `def ifloor: floor | tostring | tonumber;`, but note that this function does not work with large floating-point numbers and also loses the precision of large integers.
- gojq fixes various bugs of jq. gojq correctly deletes elements of arrays by `|= empty` ([jq#2051](https://github.com/jqlang/jq/issues/2051)). gojq fixes `try`/`catch` handling ([jq#1859](https://github.com/jqlang/jq/issues/1859), [jq#1885](https://github.com/jqlang/jq/issues/1885), [jq#2140](https://github.com/jqlang/jq/issues/2140)). gojq fixes `nth/2` to output nothing when the count is equal to or larger than the stream size ([jq#1867](https://github.com/jqlang/jq/issues/1867)). gojq consistently counts by characters (not by bytes) in `index`, `rindex`, and `indices` functions; `"１２３４５" | .[index("３"):]` results in `"３４５"` ([jq#1430](https://github.com/jqlang/jq/issues/1430), [jq#1624](https://github.com/jqlang/jq/issues/1624)). gojq handles overlapping occurrence differently in `rindex` and `indices`; `"ababa" | [rindex("aba"), indices("aba")]` results in `[2,[0,2]]` ([jq#2433](https://github.com/jqlang/jq/issues/2433)). gojq supports string indexing; `"abcde"[2]` ([jq#1520](https://github.com/jqlang/jq/issues/1520)). gojq accepts indexing query `.e0` ([jq#1526](https://github.com/jqlang/jq/issues/1526), [jq#1651](https://github.com/jqlang/jq/issues/1651)), and allows `gsub` to handle patterns including `"^"` ([jq#2148](https://github.com/jqlang/jq/issues/2148)). gojq improves variable lexer to allow using keywords for variable names, especially in binding patterns, also disallows spaces after `$` ([jq#526](https://github.com/jqlang/jq/issues/526)). gojq fixes handling files with no newline characters at the end ([jq#2374](https://github.com/jqlang/jq/issues/2374)).
- gojq truncates down floating-point numbers on indexing (`[0] | .[0.5]` results in `0` not `null`), and slicing (`[0,1,2] | .[0.5:1.5]` results in `[0]` not `[0,1]`). gojq parses unary operators with higher precedence than variable binding (`[-1 as $x | 1,$x]` results in `[1,-1]` not `[-1,-1]`). gojq implements `@uri` to escape all the reserved characters defined in RFC 3986, Sec. 2.2 ([jq#1506](https://github.com/jqlang/jq/issues/1506)), and fixes `@base64d` to allow binary string as the decoded string ([jq#1931](https://github.com/jqlang/jq/issues/1931)). gojq improves time formatting and parsing; deals with `%f` in `strftime` and `strptime` ([jq#1409](https://github.com/jqlang/jq/issues/1409)), parses timezone offsets with `fromdate` and `fromdateiso8601` ([jq#1053](https://github.com/jqlang/jq/issues/1053)), supports timezone name/offset with `%Z`/`%z` in `strptime` ([jq#929](https://github.com/jqlang/jq/issues/929), [jq#2195](https://github.com/jqlang/jq/issues/2195)), and looks up correct timezone during daylight saving time on formatting with `%Z` ([jq#1912](https://github.com/jqlang/jq/issues/1912)). gojq supports nanoseconds in date and time functions.
- gojq does not support some functions intentionally; `get_jq_origin`, `get_prog_origin`, `get_search_list` (unstable, not listed in jq document), `input_line_number`, `$__loc__` (performance issue), `recurse_down` (deprecated in jq). gojq does not support some flags; `--ascii-output, -a` (performance issue), `--seq` (not used commonly), `--sort-keys, -S` (sorts by default because `map[string]any` does not keep the order), `--unbuffered` (unbuffered by default). gojq does not parse JSON extensions supported by jq; `NaN`, `Infinity`, and `[000]`. gojq normalizes floating-point numbers to fit to double-precision floating-point numbers. gojq does not support or behaves differently with some regular expression metacharacters and flags (regular expression engine differences). gojq does not support BOM (`encoding/json` does not support this). gojq disallows using keywords for function names (`def true: .; true` is a confusing query), and module name prefixes in function declarations (using module prefixes like `def m::f: .;` is undocumented).
- gojq supports reading from YAML input (`--yaml-input`) while jq does not. gojq also supports YAML output (`--yaml-output`). gojq supports a few filters missing in jq; `scan/2` ([jq#2207](https://github.com/jqlang/jq/pull/2207)), and `@urid` format string ([jq#2261](https://github.com/jqlang/jq/issues/2261)).

### Color configuration
The gojq command automatically disables coloring output when the output is not a tty.
To force coloring output, specify `--color-output` (`-C`) option.
When [`NO_COLOR` environment variable](https://no-color.org/) is present or `--monochrome-output` (`-M`) option is specified, gojq disables coloring output.

Use `GOJQ_COLORS` environment variable to configure individual colors.
The variable is a colon-separated list of ANSI escape sequences of `null`, `false`, `true`, numbers, strings, object keys, arrays, and objects.
The default configuration is `90:33:33:36:32:34;1`.

## Usage as a library
You can use the gojq parser and interpreter from your Go products.

```go
package main

import (
	"fmt"
	"log"

	"github.com/itchyny/gojq"
)

func main() {
	query, err := gojq.Parse(".foo | ..")
	if err != nil {
		log.Fatalln(err)
	}
	input := map[string]any{"foo": []any{1, 2, 3}}
	iter := query.Run(input) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		fmt.Printf("%#v\n", v)
	}
}
```

- Firstly, use [`gojq.Parse(string) (*Query, error)`](https://pkg.go.dev/github.com/itchyny/gojq#Parse) to get the query from a string.
- Secondly, get the result iterator
  - using [`query.Run`](https://pkg.go.dev/github.com/itchyny/gojq#Query.Run) or [`query.RunWithContext`](https://pkg.go.dev/github.com/itchyny/gojq#Query.RunWithContext)
  - or alternatively, compile the query using [`gojq.Compile`](https://pkg.go.dev/github.com/itchyny/gojq#Compile) and then [`code.Run`](https://pkg.go.dev/github.com/itchyny/gojq#Code.Run) or [`code.RunWithContext`](https://pkg.go.dev/github.com/itchyny/gojq#Code.RunWithContext). You can reuse the `*Code` against multiple inputs to avoid compilation of the same query. But for arguments of `code.Run`, do not give values sharing same data between multiple calls.
  - In either case, you cannot use custom type values as the query input. The type should be `[]any` for an array and `map[string]any` for a map (just like decoded to an `any` using the [encoding/json](https://golang.org/pkg/encoding/json/) package). You can't use `[]int` or `map[string]string`, for example. If you want to query your custom struct, marshal to JSON, unmarshal to `any` and use it as the query input.
- Thirdly, iterate through the results using [`iter.Next() (any, bool)`](https://pkg.go.dev/github.com/itchyny/gojq#Iter). The iterator can emit an error so make sure to handle it. The method returns `true` with results, and `false` when the iterator terminates.
  - The return type is not `(any, error)` because iterators can emit multiple errors and you can continue after an error. It is difficult for the iterator to tell the termination in this situation.
  - Note that the result iterator may emit infinite number of values; `repeat(0)` and `range(infinite)`. It may stuck with no output value; `def f: f; f`. Use `RunWithContext` when you want to limit the execution time.

[`gojq.Compile`](https://pkg.go.dev/github.com/itchyny/gojq#Compile) allows to configure the following compiler options.

- [`gojq.WithModuleLoader`](https://pkg.go.dev/github.com/itchyny/gojq#WithModuleLoader) allows to load modules. By default, the module feature is disabled. If you want to load modules from the file system, use [`gojq.NewModuleLoader`](https://pkg.go.dev/github.com/itchyny/gojq#NewModuleLoader).
- [`gojq.WithEnvironLoader`](https://pkg.go.dev/github.com/itchyny/gojq#WithEnvironLoader) allows to configure the environment variables referenced by `env` and `$ENV`. By default, OS environment variables are not accessible due to security reasons. You can use `gojq.WithEnvironLoader(os.Environ)` if you want.
- [`gojq.WithVariables`](https://pkg.go.dev/github.com/itchyny/gojq#WithVariables) allows to configure the variables which can be used in the query. Pass the values of the variables to [`code.Run`](https://pkg.go.dev/github.com/itchyny/gojq#Code.Run) in the same order.
- [`gojq.WithFunction`](https://pkg.go.dev/github.com/itchyny/gojq#WithFunction) allows to add a custom internal function. An internal function can return a single value (which can be an error) each invocation. To add a jq function (which may include a comma operator to emit multiple values, `empty` function, accept a filter for its argument, or call another built-in function), use `LoadInitModules` of the module loader.
- [`gojq.WithIterFunction`](https://pkg.go.dev/github.com/itchyny/gojq#WithIterFunction) allows to add a custom iterator function. An iterator function returns an iterator to emit multiple values. You cannot define both iterator and non-iterator functions of the same name (with possibly different arities). You can use [`gojq.NewIter`](https://pkg.go.dev/github.com/itchyny/gojq#NewIter) to convert values or an error to a [`gojq.Iter`](https://pkg.go.dev/github.com/itchyny/gojq#Iter).
- [`gojq.WithInputIter`](https://pkg.go.dev/github.com/itchyny/gojq#WithInputIter) allows to use `input` and `inputs` functions. By default, these functions are disabled.

## Bug Tracker
Report bug at [Issues・itchyny/gojq - GitHub](https://github.com/itchyny/gojq/issues).

## Author
itchyny (https://github.com/itchyny)

## License
This software is released under the MIT License, see LICENSE.
