package gojq

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/itchyny/timefmt-go"
)

//go:generate go run -modfile=go.dev.mod _tools/gen_builtin.go -i builtin.jq -o builtin.go
var builtinFuncDefs map[string][]*FuncDef

const (
	argcount0 = 1 << iota
	argcount1
	argcount2
	argcount3
)

type function struct {
	argcount int
	iter     bool
	callback func(any, []any) any
}

func (fn function) accept(cnt int) bool {
	return fn.argcount&(1<<cnt) != 0
}

var internalFuncs map[string]function

func init() {
	internalFuncs = map[string]function{
		"empty":          argFunc0(nil),
		"path":           argFunc1(nil),
		"env":            argFunc0(nil),
		"builtins":       argFunc0(nil),
		"input":          argFunc0(nil),
		"modulemeta":     argFunc0(nil),
		"length":         argFunc0(funcLength),
		"utf8bytelength": argFunc0(funcUtf8ByteLength),
		"keys":           argFunc0(funcKeys),
		"has":            argFunc1(funcHas),
		"to_entries":     argFunc0(funcToEntries),
		"from_entries":   argFunc0(funcFromEntries),
		"add":            argFunc0(funcAdd),
		"tonumber":       argFunc0(funcToNumber),
		"tostring":       argFunc0(funcToString),
		"type":           argFunc0(funcType),
		"reverse":        argFunc0(funcReverse),
		"contains":       argFunc1(funcContains),
		"indices":        argFunc1(funcIndices),
		"index":          argFunc1(funcIndex),
		"rindex":         argFunc1(funcRindex),
		"startswith":     argFunc1(funcStartsWith),
		"endswith":       argFunc1(funcEndsWith),
		"ltrimstr":       argFunc1(funcLtrimstr),
		"rtrimstr":       argFunc1(funcRtrimstr),
		"explode":        argFunc0(funcExplode),
		"implode":        argFunc0(funcImplode),
		"split":          {argcount1 | argcount2, false, funcSplit},
		"ascii_downcase": argFunc0(funcASCIIDowncase),
		"ascii_upcase":   argFunc0(funcASCIIUpcase),
		"tojson":         argFunc0(funcToJSON),
		"fromjson":       argFunc0(funcFromJSON),
		"format":         argFunc1(funcFormat),
		"_tohtml":        argFunc0(funcToHTML),
		"_touri":         argFunc0(funcToURI),
		"_tourid":        argFunc0(funcToURId),
		"_tocsv":         argFunc0(funcToCSV),
		"_totsv":         argFunc0(funcToTSV),
		"_tosh":          argFunc0(funcToSh),
		"_tobase64":      argFunc0(funcToBase64),
		"_tobase64d":     argFunc0(funcToBase64d),
		"_index":         argFunc2(funcIndex2),
		"_slice":         argFunc3(funcSlice),
		"_plus":          argFunc0(funcOpPlus),
		"_negate":        argFunc0(funcOpNegate),
		"_add":           argFunc2(funcOpAdd),
		"_subtract":      argFunc2(funcOpSub),
		"_multiply":      argFunc2(funcOpMul),
		"_divide":        argFunc2(funcOpDiv),
		"_modulo":        argFunc2(funcOpMod),
		"_alternative":   argFunc2(funcOpAlt),
		"_equal":         argFunc2(funcOpEq),
		"_notequal":      argFunc2(funcOpNe),
		"_greater":       argFunc2(funcOpGt),
		"_less":          argFunc2(funcOpLt),
		"_greatereq":     argFunc2(funcOpGe),
		"_lesseq":        argFunc2(funcOpLe),
		"flatten":        {argcount0 | argcount1, false, funcFlatten},
		"_range":         {argcount3, true, funcRange},
		"min":            argFunc0(funcMin),
		"_min_by":        argFunc1(funcMinBy),
		"max":            argFunc0(funcMax),
		"_max_by":        argFunc1(funcMaxBy),
		"sort":           argFunc0(funcSort),
		"_sort_by":       argFunc1(funcSortBy),
		"_group_by":      argFunc1(funcGroupBy),
		"unique":         argFunc0(funcUnique),
		"_unique_by":     argFunc1(funcUniqueBy),
		"join":           argFunc1(funcJoin),
		"sin":            mathFunc("sin", math.Sin),
		"cos":            mathFunc("cos", math.Cos),
		"tan":            mathFunc("tan", math.Tan),
		"asin":           mathFunc("asin", math.Asin),
		"acos":           mathFunc("acos", math.Acos),
		"atan":           mathFunc("atan", math.Atan),
		"sinh":           mathFunc("sinh", math.Sinh),
		"cosh":           mathFunc("cosh", math.Cosh),
		"tanh":           mathFunc("tanh", math.Tanh),
		"asinh":          mathFunc("asinh", math.Asinh),
		"acosh":          mathFunc("acosh", math.Acosh),
		"atanh":          mathFunc("atanh", math.Atanh),
		"floor":          mathFunc("floor", math.Floor),
		"round":          mathFunc("round", math.Round),
		"nearbyint":      mathFunc("nearbyint", math.Round),
		"rint":           mathFunc("rint", math.Round),
		"ceil":           mathFunc("ceil", math.Ceil),
		"trunc":          mathFunc("trunc", math.Trunc),
		"significand":    mathFunc("significand", funcSignificand),
		"fabs":           mathFunc("fabs", math.Abs),
		"sqrt":           mathFunc("sqrt", math.Sqrt),
		"cbrt":           mathFunc("cbrt", math.Cbrt),
		"exp":            mathFunc("exp", math.Exp),
		"exp10":          mathFunc("exp10", funcExp10),
		"exp2":           mathFunc("exp2", math.Exp2),
		"expm1":          mathFunc("expm1", math.Expm1),
		"frexp":          argFunc0(funcFrexp),
		"modf":           argFunc0(funcModf),
		"log":            mathFunc("log", math.Log),
		"log10":          mathFunc("log10", math.Log10),
		"log1p":          mathFunc("log1p", math.Log1p),
		"log2":           mathFunc("log2", math.Log2),
		"logb":           mathFunc("logb", math.Logb),
		"gamma":          mathFunc("gamma", math.Gamma),
		"tgamma":         mathFunc("tgamma", math.Gamma),
		"lgamma":         mathFunc("lgamma", funcLgamma),
		"erf":            mathFunc("erf", math.Erf),
		"erfc":           mathFunc("erfc", math.Erfc),
		"j0":             mathFunc("j0", math.J0),
		"j1":             mathFunc("j1", math.J1),
		"y0":             mathFunc("y0", math.Y0),
		"y1":             mathFunc("y1", math.Y1),
		"atan2":          mathFunc2("atan2", math.Atan2),
		"copysign":       mathFunc2("copysign", math.Copysign),
		"drem":           mathFunc2("drem", funcDrem),
		"fdim":           mathFunc2("fdim", math.Dim),
		"fmax":           mathFunc2("fmax", math.Max),
		"fmin":           mathFunc2("fmin", math.Min),
		"fmod":           mathFunc2("fmod", math.Mod),
		"hypot":          mathFunc2("hypot", math.Hypot),
		"jn":             mathFunc2("jn", funcJn),
		"ldexp":          mathFunc2("ldexp", funcLdexp),
		"nextafter":      mathFunc2("nextafter", math.Nextafter),
		"nexttoward":     mathFunc2("nexttoward", math.Nextafter),
		"remainder":      mathFunc2("remainder", math.Remainder),
		"scalb":          mathFunc2("scalb", funcScalb),
		"scalbln":        mathFunc2("scalbln", funcScalbln),
		"yn":             mathFunc2("yn", funcYn),
		"pow":            mathFunc2("pow", math.Pow),
		"pow10":          mathFunc("pow10", funcExp10),
		"fma":            mathFunc3("fma", math.FMA),
		"infinite":       argFunc0(funcInfinite),
		"isfinite":       argFunc0(funcIsfinite),
		"isinfinite":     argFunc0(funcIsinfinite),
		"nan":            argFunc0(funcNan),
		"isnan":          argFunc0(funcIsnan),
		"isnormal":       argFunc0(funcIsnormal),
		"setpath":        argFunc2(funcSetpath),
		"delpaths":       argFunc1(funcDelpaths),
		"getpath":        argFunc1(funcGetpath),
		"transpose":      argFunc0(funcTranspose),
		"bsearch":        argFunc1(funcBsearch),
		"gmtime":         argFunc0(funcGmtime),
		"localtime":      argFunc0(funcLocaltime),
		"mktime":         argFunc0(funcMktime),
		"strftime":       argFunc1(funcStrftime),
		"strflocaltime":  argFunc1(funcStrflocaltime),
		"strptime":       argFunc1(funcStrptime),
		"now":            argFunc0(funcNow),
		"_match":         argFunc3(funcMatch),
		"_capture":       argFunc0(funcCapture),
		"error":          {argcount0 | argcount1, false, funcError},
		"halt":           argFunc0(funcHalt),
		"halt_error":     {argcount0 | argcount1, false, funcHaltError},
	}
}

func argFunc0(f func(any) any) function {
	return function{
		argcount0, false, func(v any, _ []any) any {
			return f(v)
		},
	}
}

func argFunc1(f func(_, _ any) any) function {
	return function{
		argcount1, false, func(v any, args []any) any {
			return f(v, args[0])
		},
	}
}

func argFunc2(f func(_, _, _ any) any) function {
	return function{
		argcount2, false, func(v any, args []any) any {
			return f(v, args[0], args[1])
		},
	}
}

func argFunc3(f func(_, _, _, _ any) any) function {
	return function{
		argcount3, false, func(v any, args []any) any {
			return f(v, args[0], args[1], args[2])
		},
	}
}

func mathFunc(name string, f func(float64) float64) function {
	return argFunc0(func(v any) any {
		x, ok := toFloat(v)
		if !ok {
			return &func0TypeError{name, v}
		}
		return f(x)
	})
}

func mathFunc2(name string, f func(_, _ float64) float64) function {
	return argFunc2(func(_, x, y any) any {
		l, ok := toFloat(x)
		if !ok {
			return &func0TypeError{name, x}
		}
		r, ok := toFloat(y)
		if !ok {
			return &func0TypeError{name, y}
		}
		return f(l, r)
	})
}

func mathFunc3(name string, f func(_, _, _ float64) float64) function {
	return argFunc3(func(_, a, b, c any) any {
		x, ok := toFloat(a)
		if !ok {
			return &func0TypeError{name, a}
		}
		y, ok := toFloat(b)
		if !ok {
			return &func0TypeError{name, b}
		}
		z, ok := toFloat(c)
		if !ok {
			return &func0TypeError{name, c}
		}
		return f(x, y, z)
	})
}

func funcLength(v any) any {
	switch v := v.(type) {
	case nil:
		return 0
	case int:
		if v >= 0 {
			return v
		}
		return -v
	case float64:
		return math.Abs(v)
	case *big.Int:
		if v.Sign() >= 0 {
			return v
		}
		return new(big.Int).Abs(v)
	case string:
		return len([]rune(v))
	case []any:
		return len(v)
	case map[string]any:
		return len(v)
	default:
		return &func0TypeError{"length", v}
	}
}

func funcUtf8ByteLength(v any) any {
	s, ok := v.(string)
	if !ok {
		return &func0TypeError{"utf8bytelength", v}
	}
	return len(s)
}

func funcKeys(v any) any {
	switch v := v.(type) {
	case []any:
		w := make([]any, len(v))
		for i := range v {
			w[i] = i
		}
		return w
	case map[string]any:
		w := make([]any, len(v))
		for i, k := range keys(v) {
			w[i] = k
		}
		return w
	default:
		return &func0TypeError{"keys", v}
	}
}

func keys(v map[string]any) []string {
	w := make([]string, len(v))
	var i int
	for k := range v {
		w[i] = k
		i++
	}
	sort.Strings(w)
	return w
}

func values(v any) ([]any, bool) {
	switch v := v.(type) {
	case []any:
		return v, true
	case map[string]any:
		vs := make([]any, len(v))
		for i, k := range keys(v) {
			vs[i] = v[k]
		}
		return vs, true
	default:
		return nil, false
	}
}

func funcHas(v, x any) any {
	switch v := v.(type) {
	case []any:
		if x, ok := toInt(x); ok {
			return 0 <= x && x < len(v)
		}
	case map[string]any:
		if x, ok := x.(string); ok {
			_, ok := v[x]
			return ok
		}
	case nil:
		return false
	}
	return &func1TypeError{"has", v, x}
}

func funcToEntries(v any) any {
	switch v := v.(type) {
	case []any:
		w := make([]any, len(v))
		for i, x := range v {
			w[i] = map[string]any{"key": i, "value": x}
		}
		return w
	case map[string]any:
		w := make([]any, len(v))
		for i, k := range keys(v) {
			w[i] = map[string]any{"key": k, "value": v[k]}
		}
		return w
	default:
		return &func0TypeError{"to_entries", v}
	}
}

func funcFromEntries(v any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func0TypeError{"from_entries", v}
	}
	w := make(map[string]any, len(vs))
	for _, v := range vs {
		switch v := v.(type) {
		case map[string]any:
			var (
				key   string
				value any
				ok    bool
			)
			for _, k := range [4]string{"key", "Key", "name", "Name"} {
				if k := v[k]; k != nil && k != false {
					if key, ok = k.(string); !ok {
						return &func0WrapError{"from_entries", vs, &objectKeyNotStringError{k}}
					}
					break
				}
			}
			if !ok {
				return &func0WrapError{"from_entries", vs, &objectKeyNotStringError{nil}}
			}
			for _, k := range [2]string{"value", "Value"} {
				if value, ok = v[k]; ok {
					break
				}
			}
			w[key] = value
		default:
			return &func0TypeError{"from_entries", v}
		}
	}
	return w
}

func funcAdd(v any) any {
	vs, ok := values(v)
	if !ok {
		return &func0TypeError{"add", v}
	}
	v = nil
	for _, x := range vs {
		switch x := x.(type) {
		case nil:
			continue
		case string:
			switch w := v.(type) {
			case nil:
				var sb strings.Builder
				sb.WriteString(x)
				v = &sb
				continue
			case *strings.Builder:
				w.WriteString(x)
				continue
			}
		case []any:
			switch w := v.(type) {
			case nil:
				s := make([]any, len(x))
				copy(s, x)
				v = s
				continue
			case []any:
				v = append(w, x...)
				continue
			}
		case map[string]any:
			switch w := v.(type) {
			case nil:
				m := make(map[string]any, len(x))
				for k, e := range x {
					m[k] = e
				}
				v = m
				continue
			case map[string]any:
				for k, e := range x {
					w[k] = e
				}
				continue
			}
		}
		if sb, ok := v.(*strings.Builder); ok {
			v = sb.String()
		}
		v = funcOpAdd(nil, v, x)
		if err, ok := v.(error); ok {
			return err
		}
	}
	if sb, ok := v.(*strings.Builder); ok {
		v = sb.String()
	}
	return v
}

func funcToNumber(v any) any {
	switch v := v.(type) {
	case int, float64, *big.Int:
		return v
	case string:
		if !newLexer(v).validNumber() {
			return &func0WrapError{"tonumber", v, errors.New("invalid number")}
		}
		return toNumber(v)
	default:
		return &func0TypeError{"tonumber", v}
	}
}

func toNumber(v string) any {
	return normalizeNumber(json.Number(v))
}

func funcToString(v any) any {
	if s, ok := v.(string); ok {
		return s
	}
	return funcToJSON(v)
}

func funcType(v any) any {
	return TypeOf(v)
}

func funcReverse(v any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func0TypeError{"reverse", v}
	}
	ws := make([]any, len(vs))
	for i, v := range vs {
		ws[len(ws)-i-1] = v
	}
	return ws
}

func funcContains(v, x any) any {
	return binopTypeSwitch(v, x,
		func(l, r int) any { return l == r },
		func(l, r float64) any { return l == r },
		func(l, r *big.Int) any { return l.Cmp(r) == 0 },
		func(l, r string) any { return strings.Contains(l, r) },
		func(l, r []any) any {
		R:
			for _, r := range r {
				for _, l := range l {
					if funcContains(l, r) == true {
						continue R
					}
				}
				return false
			}
			return true
		},
		func(l, r map[string]any) any {
			if len(l) < len(r) {
				return false
			}
			for k, r := range r {
				if l, ok := l[k]; !ok || funcContains(l, r) != true {
					return false
				}
			}
			return true
		},
		func(l, r any) any {
			if l == r {
				return true
			}
			return &func1TypeError{"contains", l, r}
		},
	)
}

func funcIndices(v, x any) any {
	return indexFunc("indices", v, x, indices)
}

func indices(vs, xs []any) any {
	rs := []any{}
	if len(xs) == 0 {
		return rs
	}
	for i := 0; i <= len(vs)-len(xs); i++ {
		if compare(vs[i:i+len(xs)], xs) == 0 {
			rs = append(rs, i)
		}
	}
	return rs
}

func funcIndex(v, x any) any {
	return indexFunc("index", v, x, func(vs, xs []any) any {
		if len(xs) == 0 {
			return nil
		}
		for i := 0; i <= len(vs)-len(xs); i++ {
			if compare(vs[i:i+len(xs)], xs) == 0 {
				return i
			}
		}
		return nil
	})
}

func funcRindex(v, x any) any {
	return indexFunc("rindex", v, x, func(vs, xs []any) any {
		if len(xs) == 0 {
			return nil
		}
		for i := len(vs) - len(xs); i >= 0; i-- {
			if compare(vs[i:i+len(xs)], xs) == 0 {
				return i
			}
		}
		return nil
	})
}

func indexFunc(name string, v, x any, f func(_, _ []any) any) any {
	switch v := v.(type) {
	case nil:
		return nil
	case []any:
		switch x := x.(type) {
		case []any:
			return f(v, x)
		default:
			return f(v, []any{x})
		}
	case string:
		if x, ok := x.(string); ok {
			return f(explode(v), explode(x))
		}
		return &func1TypeError{name, v, x}
	default:
		return &func1TypeError{name, v, x}
	}
}

func funcStartsWith(v, x any) any {
	s, ok := v.(string)
	if !ok {
		return &func1TypeError{"startswith", v, x}
	}
	t, ok := x.(string)
	if !ok {
		return &func1TypeError{"startswith", v, x}
	}
	return strings.HasPrefix(s, t)
}

func funcEndsWith(v, x any) any {
	s, ok := v.(string)
	if !ok {
		return &func1TypeError{"endswith", v, x}
	}
	t, ok := x.(string)
	if !ok {
		return &func1TypeError{"endswith", v, x}
	}
	return strings.HasSuffix(s, t)
}

func funcLtrimstr(v, x any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	t, ok := x.(string)
	if !ok {
		return v
	}
	return strings.TrimPrefix(s, t)
}

func funcRtrimstr(v, x any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	t, ok := x.(string)
	if !ok {
		return v
	}
	return strings.TrimSuffix(s, t)
}

func funcExplode(v any) any {
	s, ok := v.(string)
	if !ok {
		return &func0TypeError{"explode", v}
	}
	return explode(s)
}

func explode(s string) []any {
	xs := make([]any, len([]rune(s)))
	var i int
	for _, r := range s {
		xs[i] = int(r)
		i++
	}
	return xs
}

func funcImplode(v any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func0TypeError{"implode", v}
	}
	var sb strings.Builder
	sb.Grow(len(vs))
	for _, v := range vs {
		if r, ok := toInt(v); ok && 0 <= r && r <= utf8.MaxRune {
			sb.WriteRune(rune(r))
		} else {
			return &func0TypeError{"implode", vs}
		}
	}
	return sb.String()
}

func funcSplit(v any, args []any) any {
	s, ok := v.(string)
	if !ok {
		return &func0TypeError{"split", v}
	}
	x, ok := args[0].(string)
	if !ok {
		return &func0TypeError{"split", x}
	}
	var ss []string
	if len(args) == 1 {
		ss = strings.Split(s, x)
	} else {
		var flags string
		if args[1] != nil {
			v, ok := args[1].(string)
			if !ok {
				return &func0TypeError{"split", args[1]}
			}
			flags = v
		}
		r, err := compileRegexp(x, flags)
		if err != nil {
			return err
		}
		ss = r.Split(s, -1)
	}
	xs := make([]any, len(ss))
	for i, s := range ss {
		xs[i] = s
	}
	return xs
}

func funcASCIIDowncase(v any) any {
	s, ok := v.(string)
	if !ok {
		return &func0TypeError{"ascii_downcase", v}
	}
	return strings.Map(func(r rune) rune {
		if 'A' <= r && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, s)
}

func funcASCIIUpcase(v any) any {
	s, ok := v.(string)
	if !ok {
		return &func0TypeError{"ascii_upcase", v}
	}
	return strings.Map(func(r rune) rune {
		if 'a' <= r && r <= 'z' {
			return r - ('a' - 'A')
		}
		return r
	}, s)
}

func funcToJSON(v any) any {
	return jsonMarshal(v)
}

func funcFromJSON(v any) any {
	s, ok := v.(string)
	if !ok {
		return &func0TypeError{"fromjson", v}
	}
	var w any
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	if err := dec.Decode(&w); err != nil {
		return &func0WrapError{"fromjson", v, err}
	}
	if _, err := dec.Token(); err != io.EOF {
		return &func0TypeError{"fromjson", v}
	}
	return normalizeNumbers(w)
}

func funcFormat(v, x any) any {
	s, ok := x.(string)
	if !ok {
		return &func0TypeError{"format", x}
	}
	format := "@" + s
	f := formatToFunc(format)
	if f == nil {
		return &formatNotFoundError{format}
	}
	return internalFuncs[f.Name].callback(v, nil)
}

var htmlEscaper = strings.NewReplacer(
	`<`, "&lt;",
	`>`, "&gt;",
	`&`, "&amp;",
	`'`, "&apos;",
	`"`, "&quot;",
)

func funcToHTML(v any) any {
	switch x := funcToString(v).(type) {
	case string:
		return htmlEscaper.Replace(x)
	default:
		return x
	}
}

func funcToURI(v any) any {
	switch x := funcToString(v).(type) {
	case string:
		return url.QueryEscape(x)
	default:
		return x
	}
}

func funcToURId(v any) any {
	switch x := funcToString(v).(type) {
	case string:
		x, err := url.QueryUnescape(x)
		if err != nil {
			return &func0WrapError{"@urid", v, err}
		}
		return x
	default:
		return x
	}
}

var csvEscaper = strings.NewReplacer(
	`"`, `""`,
	"\x00", `\0`,
)

func funcToCSV(v any) any {
	return formatJoin("csv", v, ",", func(s string) string {
		return `"` + csvEscaper.Replace(s) + `"`
	})
}

var tsvEscaper = strings.NewReplacer(
	"\t", `\t`,
	"\r", `\r`,
	"\n", `\n`,
	"\\", `\\`,
	"\x00", `\0`,
)

func funcToTSV(v any) any {
	return formatJoin("tsv", v, "\t", tsvEscaper.Replace)
}

var shEscaper = strings.NewReplacer(
	"'", `'\''`,
	"\x00", `\0`,
)

func funcToSh(v any) any {
	if _, ok := v.([]any); !ok {
		v = []any{v}
	}
	return formatJoin("sh", v, " ", func(s string) string {
		return "'" + shEscaper.Replace(s) + "'"
	})
}

func formatJoin(typ string, v any, sep string, escape func(string) string) any {
	vs, ok := v.([]any)
	if !ok {
		return &func0TypeError{"@" + typ, v}
	}
	ss := make([]string, len(vs))
	for i, v := range vs {
		switch v := v.(type) {
		case []any, map[string]any:
			return &formatRowError{typ, v}
		case string:
			ss[i] = escape(v)
		default:
			if s := jsonMarshal(v); s != "null" || typ == "sh" {
				ss[i] = s
			}
		}
	}
	return strings.Join(ss, sep)
}

func funcToBase64(v any) any {
	switch x := funcToString(v).(type) {
	case string:
		return base64.StdEncoding.EncodeToString([]byte(x))
	default:
		return x
	}
}

func funcToBase64d(v any) any {
	switch x := funcToString(v).(type) {
	case string:
		if i := strings.IndexRune(x, base64.StdPadding); i >= 0 {
			x = x[:i]
		}
		y, err := base64.RawStdEncoding.DecodeString(x)
		if err != nil {
			return &func0WrapError{"@base64d", v, err}
		}
		return string(y)
	default:
		return x
	}
}

func funcIndex2(_, v, x any) any {
	switch x := x.(type) {
	case string:
		switch v := v.(type) {
		case nil:
			return nil
		case map[string]any:
			return v[x]
		default:
			return &expectedObjectError{v}
		}
	case int, float64, *big.Int:
		i, _ := toInt(x)
		switch v := v.(type) {
		case nil:
			return nil
		case []any:
			return index(v, i)
		case string:
			return indexString(v, i)
		default:
			return &expectedArrayError{v}
		}
	case []any:
		switch v := v.(type) {
		case nil:
			return nil
		case []any:
			return indices(v, x)
		default:
			return &expectedArrayError{v}
		}
	case map[string]any:
		if v == nil {
			return nil
		}
		start, ok := x["start"]
		if !ok {
			return &expectedStartEndError{x}
		}
		end, ok := x["end"]
		if !ok {
			return &expectedStartEndError{x}
		}
		return funcSlice(nil, v, end, start)
	default:
		switch v.(type) {
		case []any:
			return &arrayIndexNotNumberError{x}
		case string:
			return &stringIndexNotNumberError{x}
		default:
			return &objectKeyNotStringError{x}
		}
	}
}

func index(vs []any, i int) any {
	i = clampIndex(i, -1, len(vs))
	if 0 <= i && i < len(vs) {
		return vs[i]
	}
	return nil
}

func indexString(s string, i int) any {
	l := len([]rune(s))
	i = clampIndex(i, -1, l)
	if 0 <= i && i < l {
		for _, r := range s {
			if i--; i < 0 {
				return string(r)
			}
		}
	}
	return nil
}

func funcSlice(_, v, e, s any) (r any) {
	switch v := v.(type) {
	case nil:
		return nil
	case []any:
		return slice(v, e, s)
	case string:
		return sliceString(v, e, s)
	default:
		return &expectedArrayError{v}
	}
}

func slice(vs []any, e, s any) any {
	var start, end int
	if s != nil {
		if i, ok := toInt(s); ok {
			start = clampIndex(i, 0, len(vs))
		} else {
			return &arrayIndexNotNumberError{s}
		}
	}
	if e != nil {
		if i, ok := toInt(e); ok {
			end = clampIndex(i, start, len(vs))
		} else {
			return &arrayIndexNotNumberError{e}
		}
	} else {
		end = len(vs)
	}
	return vs[start:end]
}

func sliceString(v string, e, s any) any {
	var start, end int
	l := len([]rune(v))
	if s != nil {
		if i, ok := toInt(s); ok {
			start = clampIndex(i, 0, l)
		} else {
			return &stringIndexNotNumberError{s}
		}
	}
	if e != nil {
		if i, ok := toInt(e); ok {
			end = clampIndex(i, start, l)
		} else {
			return &stringIndexNotNumberError{e}
		}
	} else {
		end = l
	}
	if start < l {
		for i := range v {
			if start--; start < 0 {
				start = i
				break
			}
		}
	} else {
		start = len(v)
	}
	if end < l {
		for i := range v {
			if end--; end < 0 {
				end = i
				break
			}
		}
	} else {
		end = len(v)
	}
	return v[start:end]
}

func clampIndex(i, min, max int) int {
	if i < 0 {
		i += max
	}
	if i < min {
		return min
	} else if i < max {
		return i
	} else {
		return max
	}
}

func funcFlatten(v any, args []any) any {
	vs, ok := values(v)
	if !ok {
		return &func0TypeError{"flatten", v}
	}
	var depth float64
	if len(args) == 0 {
		depth = -1
	} else {
		depth, ok = toFloat(args[0])
		if !ok {
			return &func0TypeError{"flatten", args[0]}
		}
		if depth < 0 {
			return &flattenDepthError{depth}
		}
	}
	return flatten([]any{}, vs, depth)
}

func flatten(xs, vs []any, depth float64) []any {
	for _, v := range vs {
		if vs, ok := v.([]any); ok && depth != 0 {
			xs = flatten(xs, vs, depth-1)
		} else {
			xs = append(xs, v)
		}
	}
	return xs
}

type rangeIter struct {
	value, end, step any
}

func (iter *rangeIter) Next() (any, bool) {
	if compare(iter.step, 0)*compare(iter.value, iter.end) >= 0 {
		return nil, false
	}
	v := iter.value
	iter.value = funcOpAdd(nil, v, iter.step)
	return v, true
}

func funcRange(_ any, xs []any) any {
	for _, x := range xs {
		switch x.(type) {
		case int, float64, *big.Int:
		default:
			return &func0TypeError{"range", x}
		}
	}
	return &rangeIter{xs[0], xs[1], xs[2]}
}

func funcMin(v any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func0TypeError{"min", v}
	}
	return minMaxBy(vs, vs, true)
}

func funcMinBy(v, x any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func1TypeError{"min_by", v, x}
	}
	xs, ok := x.([]any)
	if !ok {
		return &func1TypeError{"min_by", v, x}
	}
	if len(vs) != len(xs) {
		return &func1WrapError{"min_by", v, x, &lengthMismatchError{}}
	}
	return minMaxBy(vs, xs, true)
}

func funcMax(v any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func0TypeError{"max", v}
	}
	return minMaxBy(vs, vs, false)
}

func funcMaxBy(v, x any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func1TypeError{"max_by", v, x}
	}
	xs, ok := x.([]any)
	if !ok {
		return &func1TypeError{"max_by", v, x}
	}
	if len(vs) != len(xs) {
		return &func1WrapError{"max_by", v, x, &lengthMismatchError{}}
	}
	return minMaxBy(vs, xs, false)
}

func minMaxBy(vs, xs []any, isMin bool) any {
	if len(vs) == 0 {
		return nil
	}
	i, j, x := 0, 0, xs[0]
	for i++; i < len(xs); i++ {
		if compare(x, xs[i]) > 0 == isMin {
			j, x = i, xs[i]
		}
	}
	return vs[j]
}

type sortItem struct {
	value, key any
}

func sortItems(name string, v, x any) ([]*sortItem, error) {
	vs, ok := v.([]any)
	if !ok {
		if strings.HasSuffix(name, "_by") {
			return nil, &func1TypeError{name, v, x}
		}
		return nil, &func0TypeError{name, v}
	}
	xs, ok := x.([]any)
	if !ok {
		return nil, &func1TypeError{name, v, x}
	}
	if len(vs) != len(xs) {
		return nil, &func1WrapError{name, v, x, &lengthMismatchError{}}
	}
	items := make([]*sortItem, len(vs))
	for i, v := range vs {
		items[i] = &sortItem{v, xs[i]}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return compare(items[i].key, items[j].key) < 0
	})
	return items, nil
}

func funcSort(v any) any {
	return sortBy("sort", v, v)
}

func funcSortBy(v, x any) any {
	return sortBy("sort_by", v, x)
}

func sortBy(name string, v, x any) any {
	items, err := sortItems(name, v, x)
	if err != nil {
		return err
	}
	rs := make([]any, len(items))
	for i, x := range items {
		rs[i] = x.value
	}
	return rs
}

func funcGroupBy(v, x any) any {
	items, err := sortItems("group_by", v, x)
	if err != nil {
		return err
	}
	rs := []any{}
	var last any
	for i, r := range items {
		if i == 0 || compare(last, r.key) != 0 {
			rs, last = append(rs, []any{r.value}), r.key
		} else {
			rs[len(rs)-1] = append(rs[len(rs)-1].([]any), r.value)
		}
	}
	return rs
}

func funcUnique(v any) any {
	return uniqueBy("unique", v, v)
}

func funcUniqueBy(v, x any) any {
	return uniqueBy("unique_by", v, x)
}

func uniqueBy(name string, v, x any) any {
	items, err := sortItems(name, v, x)
	if err != nil {
		return err
	}
	rs := []any{}
	var last any
	for i, r := range items {
		if i == 0 || compare(last, r.key) != 0 {
			rs, last = append(rs, r.value), r.key
		}
	}
	return rs
}

func funcJoin(v, x any) any {
	vs, ok := values(v)
	if !ok {
		return &func1TypeError{"join", v, x}
	}
	if len(vs) == 0 {
		return ""
	}
	sep, ok := x.(string)
	if len(vs) > 1 && !ok {
		return &func1TypeError{"join", v, x}
	}
	ss := make([]string, len(vs))
	for i, v := range vs {
		switch v := v.(type) {
		case nil:
		case string:
			ss[i] = v
		case bool:
			if v {
				ss[i] = "true"
			} else {
				ss[i] = "false"
			}
		case int, float64, *big.Int:
			ss[i] = jsonMarshal(v)
		default:
			return &joinTypeError{v}
		}
	}
	return strings.Join(ss, sep)
}

func funcSignificand(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0.0 {
		return v
	}
	return math.Float64frombits((math.Float64bits(v) & 0x800fffffffffffff) | 0x3ff0000000000000)
}

func funcExp10(v float64) float64 {
	return math.Pow(10, v)
}

func funcFrexp(v any) any {
	x, ok := toFloat(v)
	if !ok {
		return &func0TypeError{"frexp", v}
	}
	f, e := math.Frexp(x)
	return []any{f, e}
}

func funcModf(v any) any {
	x, ok := toFloat(v)
	if !ok {
		return &func0TypeError{"modf", v}
	}
	i, f := math.Modf(x)
	return []any{f, i}
}

func funcLgamma(v float64) float64 {
	v, _ = math.Lgamma(v)
	return v
}

func funcDrem(l, r float64) float64 {
	x := math.Remainder(l, r)
	if x == 0.0 {
		return math.Copysign(x, l)
	}
	return x
}

func funcJn(l, r float64) float64 {
	return math.Jn(int(l), r)
}

func funcLdexp(l, r float64) float64 {
	return math.Ldexp(l, int(r))
}

func funcScalb(l, r float64) float64 {
	return l * math.Pow(2, r)
}

func funcScalbln(l, r float64) float64 {
	return l * math.Pow(2, r)
}

func funcYn(l, r float64) float64 {
	return math.Yn(int(l), r)
}

func funcInfinite(any) any {
	return math.Inf(1)
}

func funcIsfinite(v any) any {
	x, ok := toFloat(v)
	return ok && !math.IsInf(x, 0)
}

func funcIsinfinite(v any) any {
	x, ok := toFloat(v)
	return ok && math.IsInf(x, 0)
}

func funcNan(any) any {
	return math.NaN()
}

func funcIsnan(v any) any {
	x, ok := toFloat(v)
	if !ok {
		if v == nil {
			return false
		}
		return &func0TypeError{"isnan", v}
	}
	return math.IsNaN(x)
}

func funcIsnormal(v any) any {
	if v, ok := toFloat(v); ok {
		e := math.Float64bits(v) & 0x7ff0000000000000 >> 52
		return 0 < e && e < 0x7ff
	}
	return false
}

// An `allocator` creates new maps and slices, stores the allocated addresses.
// This allocator is used to reduce allocations on assignment operator (`=`),
// update-assignment operator (`|=`), and the `map_values`, `del`, `delpaths`
// functions.
type allocator map[uintptr]struct{}

func funcAllocator(any, []any) any {
	return allocator{}
}

func (a allocator) allocated(v any) bool {
	_, ok := a[reflect.ValueOf(v).Pointer()]
	return ok
}

func (a allocator) makeObject(l int) map[string]any {
	v := make(map[string]any, l)
	if a != nil {
		a[reflect.ValueOf(v).Pointer()] = struct{}{}
	}
	return v
}

func (a allocator) makeArray(l, c int) []any {
	if c < l {
		c = l
	}
	v := make([]any, l, c)
	if a != nil {
		a[reflect.ValueOf(v).Pointer()] = struct{}{}
	}
	return v
}

func funcSetpath(v, p, n any) any {
	// There is no need to use an allocator on a single update.
	return setpath(v, p, n, nil)
}

// Used in compiler#compileAssign and compiler#compileModify.
func funcSetpathWithAllocator(v any, args []any) any {
	return setpath(v, args[0], args[1], args[2].(allocator))
}

func setpath(v, p, n any, a allocator) any {
	path, ok := p.([]any)
	if !ok {
		return &func1TypeError{"setpath", v, p}
	}
	u, err := update(v, path, n, a)
	if err != nil {
		return &func2WrapError{"setpath", v, p, n, err}
	}
	return u
}

func funcDelpaths(v, p any) any {
	return delpaths(v, p, allocator{})
}

// Used in compiler#compileAssign and compiler#compileModify.
func funcDelpathsWithAllocator(v any, args []any) any {
	return delpaths(v, args[0], args[1].(allocator))
}

func delpaths(v, p any, a allocator) any {
	paths, ok := p.([]any)
	if !ok {
		return &func1TypeError{"delpaths", v, p}
	}
	if len(paths) == 0 {
		return v
	}
	// Fills the paths with an empty value and then delete them. We cannot delete
	// in each loop because array indices should not change. For example,
	//   jq -n "[0, 1, 2, 3] | delpaths([[1], [2]])" #=> [0, 3].
	var empty struct{}
	var err error
	u := v
	for _, q := range paths {
		path, ok := q.([]any)
		if !ok {
			return &func1WrapError{"delpaths", v, p, &expectedArrayError{q}}
		}
		u, err = update(u, path, empty, a)
		if err != nil {
			return &func1WrapError{"delpaths", v, p, err}
		}
	}
	return deleteEmpty(u)
}

func update(v any, path []any, n any, a allocator) (any, error) {
	if len(path) == 0 {
		return n, nil
	}
	switch p := path[0].(type) {
	case string:
		switch v := v.(type) {
		case nil:
			return updateObject(nil, p, path[1:], n, a)
		case map[string]any:
			return updateObject(v, p, path[1:], n, a)
		case struct{}:
			return v, nil
		default:
			return nil, &expectedObjectError{v}
		}
	case int, float64, *big.Int:
		i, _ := toInt(p)
		switch v := v.(type) {
		case nil:
			return updateArrayIndex(nil, i, path[1:], n, a)
		case []any:
			return updateArrayIndex(v, i, path[1:], n, a)
		case struct{}:
			return v, nil
		default:
			return nil, &expectedArrayError{v}
		}
	case map[string]any:
		switch v := v.(type) {
		case nil:
			return updateArraySlice(nil, p, path[1:], n, a)
		case []any:
			return updateArraySlice(v, p, path[1:], n, a)
		case struct{}:
			return v, nil
		default:
			return nil, &expectedArrayError{v}
		}
	default:
		switch v.(type) {
		case []any:
			return nil, &arrayIndexNotNumberError{p}
		default:
			return nil, &objectKeyNotStringError{p}
		}
	}
}

func updateObject(v map[string]any, k string, path []any, n any, a allocator) (any, error) {
	x, ok := v[k]
	if !ok && n == struct{}{} {
		return v, nil
	}
	u, err := update(x, path, n, a)
	if err != nil {
		return nil, err
	}
	if a.allocated(v) {
		v[k] = u
		return v, nil
	}
	w := a.makeObject(len(v) + 1)
	for k, v := range v {
		w[k] = v
	}
	w[k] = u
	return w, nil
}

func updateArrayIndex(v []any, i int, path []any, n any, a allocator) (any, error) {
	var x any
	if j := clampIndex(i, -1, len(v)); j < 0 {
		if n == struct{}{} {
			return v, nil
		}
		return nil, &arrayIndexNegativeError{i}
	} else if j < len(v) {
		i = j
		x = v[i]
	} else {
		if n == struct{}{} {
			return v, nil
		}
		if i >= 0x8000000 {
			return nil, &arrayIndexTooLargeError{i}
		}
	}
	u, err := update(x, path, n, a)
	if err != nil {
		return nil, err
	}
	l, c := len(v), cap(v)
	if a.allocated(v) {
		if i < c {
			if i >= l {
				v = v[:i+1]
			}
			v[i] = u
			return v, nil
		}
		c *= 2
	}
	if i >= l {
		l = i + 1
	}
	w := a.makeArray(l, c)
	copy(w, v)
	w[i] = u
	return w, nil
}

func updateArraySlice(v []any, m map[string]any, path []any, n any, a allocator) (any, error) {
	s, ok := m["start"]
	if !ok {
		return nil, &expectedStartEndError{m}
	}
	e, ok := m["end"]
	if !ok {
		return nil, &expectedStartEndError{m}
	}
	var start, end int
	if i, ok := toInt(s); ok {
		start = clampIndex(i, 0, len(v))
	}
	if i, ok := toInt(e); ok {
		end = clampIndex(i, start, len(v))
	} else {
		end = len(v)
	}
	if start == end && n == struct{}{} {
		return v, nil
	}
	u, err := update(v[start:end], path, n, a)
	if err != nil {
		return nil, err
	}
	switch u := u.(type) {
	case []any:
		var w []any
		if len(u) == end-start && a.allocated(v) {
			w = v
		} else {
			w = a.makeArray(len(v)-(end-start)+len(u), 0)
			copy(w, v[:start])
			copy(w[start+len(u):], v[end:])
		}
		copy(w[start:], u)
		return w, nil
	case struct{}:
		var w []any
		if a.allocated(v) {
			w = v
		} else {
			w = a.makeArray(len(v), 0)
			copy(w, v)
		}
		for i := start; i < end; i++ {
			w[i] = u
		}
		return w, nil
	default:
		return nil, &expectedArrayError{u}
	}
}

func deleteEmpty(v any) any {
	switch v := v.(type) {
	case struct{}:
		return nil
	case map[string]any:
		for k, w := range v {
			if w == struct{}{} {
				delete(v, k)
			} else {
				v[k] = deleteEmpty(w)
			}
		}
		return v
	case []any:
		var j int
		for _, w := range v {
			if w != struct{}{} {
				v[j] = deleteEmpty(w)
				j++
			}
		}
		for i := j; i < len(v); i++ {
			v[i] = nil
		}
		return v[:j]
	default:
		return v
	}
}

func funcGetpath(v, p any) any {
	path, ok := p.([]any)
	if !ok {
		return &func1TypeError{"getpath", v, p}
	}
	u := v
	for _, x := range path {
		switch v.(type) {
		case nil, []any, map[string]any:
			v = funcIndex2(nil, v, x)
			if err, ok := v.(error); ok {
				return &func1WrapError{"getpath", u, p, err}
			}
		default:
			return &func1TypeError{"getpath", u, p}
		}
	}
	return v
}

func funcTranspose(v any) any {
	vss, ok := v.([]any)
	if !ok {
		return &func0TypeError{"transpose", v}
	}
	if len(vss) == 0 {
		return []any{}
	}
	var l int
	for _, vs := range vss {
		vs, ok := vs.([]any)
		if !ok {
			return &func0TypeError{"transpose", v}
		}
		if k := len(vs); l < k {
			l = k
		}
	}
	wss := make([][]any, l)
	xs := make([]any, l)
	for i, k := 0, len(vss); i < l; i++ {
		s := make([]any, k)
		wss[i] = s
		xs[i] = s
	}
	for i, vs := range vss {
		for j, v := range vs.([]any) {
			wss[j][i] = v
		}
	}
	return xs
}

func funcBsearch(v, t any) any {
	vs, ok := v.([]any)
	if !ok {
		return &func1TypeError{"bsearch", v, t}
	}
	i := sort.Search(len(vs), func(i int) bool {
		return compare(vs[i], t) >= 0
	})
	if i < len(vs) && compare(vs[i], t) == 0 {
		return i
	}
	return -i - 1
}

func funcGmtime(v any) any {
	if v, ok := toFloat(v); ok {
		return epochToArray(v, time.UTC)
	}
	return &func0TypeError{"gmtime", v}
}

func funcLocaltime(v any) any {
	if v, ok := toFloat(v); ok {
		return epochToArray(v, time.Local)
	}
	return &func0TypeError{"localtime", v}
}

func epochToArray(v float64, loc *time.Location) []any {
	t := time.Unix(int64(v), int64((v-math.Floor(v))*1e9)).In(loc)
	return []any{
		t.Year(),
		int(t.Month()) - 1,
		t.Day(),
		t.Hour(),
		t.Minute(),
		float64(t.Second()) + float64(t.Nanosecond())/1e9,
		int(t.Weekday()),
		t.YearDay() - 1,
	}
}

func funcMktime(v any) any {
	a, ok := v.([]any)
	if !ok {
		return &func0TypeError{"mktime", v}
	}
	t, err := arrayToTime(a, time.UTC)
	if err != nil {
		return &func0WrapError{"mktime", v, err}
	}
	return timeToEpoch(t)
}

func timeToEpoch(t time.Time) float64 {
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

func funcStrftime(v, x any) any {
	if w, ok := toFloat(v); ok {
		v = epochToArray(w, time.UTC)
	}
	a, ok := v.([]any)
	if !ok {
		return &func1TypeError{"strftime", v, x}
	}
	format, ok := x.(string)
	if !ok {
		return &func1TypeError{"strftime", v, x}
	}
	t, err := arrayToTime(a, time.UTC)
	if err != nil {
		return &func1WrapError{"strftime", v, x, err}
	}
	return timefmt.Format(t, format)
}

func funcStrflocaltime(v, x any) any {
	if w, ok := toFloat(v); ok {
		v = epochToArray(w, time.Local)
	}
	a, ok := v.([]any)
	if !ok {
		return &func1TypeError{"strflocaltime", v, x}
	}
	format, ok := x.(string)
	if !ok {
		return &func1TypeError{"strflocaltime", v, x}
	}
	t, err := arrayToTime(a, time.Local)
	if err != nil {
		return &func1WrapError{"strflocaltime", v, x, err}
	}
	return timefmt.Format(t, format)
}

func funcStrptime(v, x any) any {
	s, ok := v.(string)
	if !ok {
		return &func1TypeError{"strptime", v, x}
	}
	format, ok := x.(string)
	if !ok {
		return &func1TypeError{"strptime", v, x}
	}
	t, err := timefmt.Parse(s, format)
	if err != nil {
		return &func1WrapError{"strptime", v, x, err}
	}
	var u time.Time
	if t == u {
		return &func1TypeError{"strptime", v, x}
	}
	return epochToArray(timeToEpoch(t), time.UTC)
}

func arrayToTime(a []any, loc *time.Location) (time.Time, error) {
	var t time.Time
	if len(a) != 8 {
		return t, &timeArrayError{}
	}
	var y, m, d, h, min, sec, nsec int
	var ok bool
	if y, ok = toInt(a[0]); !ok {
		return t, &timeArrayError{}
	}
	if m, ok = toInt(a[1]); ok {
		m++
	} else {
		return t, &timeArrayError{}
	}
	if d, ok = toInt(a[2]); !ok {
		return t, &timeArrayError{}
	}
	if h, ok = toInt(a[3]); !ok {
		return t, &timeArrayError{}
	}
	if min, ok = toInt(a[4]); !ok {
		return t, &timeArrayError{}
	}
	if x, ok := toFloat(a[5]); ok {
		sec = int(x)
		nsec = int((x - math.Floor(x)) * 1e9)
	} else {
		return t, &timeArrayError{}
	}
	if _, ok = toFloat(a[6]); !ok {
		return t, &timeArrayError{}
	}
	if _, ok = toFloat(a[7]); !ok {
		return t, &timeArrayError{}
	}
	return time.Date(y, time.Month(m), d, h, min, sec, nsec, loc), nil
}

func funcNow(any) any {
	return timeToEpoch(time.Now())
}

func funcMatch(v, re, fs, testing any) any {
	name := "match"
	if testing == true {
		name = "test"
	}
	var flags string
	if fs != nil {
		v, ok := fs.(string)
		if !ok {
			return &func2TypeError{name, v, re, fs}
		}
		flags = v
	}
	s, ok := v.(string)
	if !ok {
		return &func2TypeError{name, v, re, fs}
	}
	restr, ok := re.(string)
	if !ok {
		return &func2TypeError{name, v, re, fs}
	}
	r, err := compileRegexp(restr, flags)
	if err != nil {
		return err
	}
	var xs [][]int
	if strings.ContainsRune(flags, 'g') && testing != true {
		xs = r.FindAllStringSubmatchIndex(s, -1)
	} else {
		got := r.FindStringSubmatchIndex(s)
		if testing == true {
			return got != nil
		}
		if got != nil {
			xs = [][]int{got}
		}
	}
	res, names := make([]any, len(xs)), r.SubexpNames()
	for i, x := range xs {
		captures := make([]any, (len(x)-2)/2)
		for j := 1; j < len(x)/2; j++ {
			var name any
			if n := names[j]; n != "" {
				name = n
			}
			if x[j*2] < 0 {
				captures[j-1] = map[string]any{
					"name":   name,
					"offset": -1,
					"length": 0,
					"string": nil,
				}
				continue
			}
			captures[j-1] = map[string]any{
				"name":   name,
				"offset": len([]rune(s[:x[j*2]])),
				"length": len([]rune(s[:x[j*2+1]])) - len([]rune(s[:x[j*2]])),
				"string": s[x[j*2]:x[j*2+1]],
			}
		}
		res[i] = map[string]any{
			"offset":   len([]rune(s[:x[0]])),
			"length":   len([]rune(s[:x[1]])) - len([]rune(s[:x[0]])),
			"string":   s[x[0]:x[1]],
			"captures": captures,
		}
	}
	return res
}

func compileRegexp(re, flags string) (*regexp.Regexp, error) {
	if strings.IndexFunc(flags, func(r rune) bool {
		return r != 'g' && r != 'i' && r != 'm'
	}) >= 0 {
		return nil, fmt.Errorf("unsupported regular expression flag: %q", flags)
	}
	re = strings.ReplaceAll(re, "(?<", "(?P<")
	if strings.ContainsRune(flags, 'i') {
		re = "(?i)" + re
	}
	if strings.ContainsRune(flags, 'm') {
		re = "(?s)" + re
	}
	r, err := regexp.Compile(re)
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression %q: %s", re, err)
	}
	return r, nil
}

func funcCapture(v any) any {
	vs, ok := v.(map[string]any)
	if !ok {
		return &expectedObjectError{v}
	}
	v = vs["captures"]
	captures, ok := v.([]any)
	if !ok {
		return &expectedArrayError{v}
	}
	w := make(map[string]any, len(captures))
	for _, capture := range captures {
		if capture, ok := capture.(map[string]any); ok {
			if name, ok := capture["name"].(string); ok {
				w[name] = capture["string"]
			}
		}
	}
	return w
}

func funcError(v any, args []any) any {
	if len(args) > 0 {
		v = args[0]
	}
	code := 5
	if v == nil {
		code = 0
	}
	return &exitCodeError{v, code, false}
}

func funcHalt(any) any {
	return &exitCodeError{nil, 0, true}
}

func funcHaltError(v any, args []any) any {
	code := 5
	if len(args) > 0 {
		var ok bool
		if code, ok = toInt(args[0]); !ok {
			return &func0TypeError{"halt_error", args[0]}
		}
	}
	return &exitCodeError{v, code, true}
}

func toInt(x any) (int, bool) {
	switch x := x.(type) {
	case int:
		return x, true
	case float64:
		return floatToInt(x), true
	case *big.Int:
		if x.IsInt64() {
			if i := x.Int64(); math.MinInt <= i && i <= math.MaxInt {
				return int(i), true
			}
		}
		if x.Sign() > 0 {
			return math.MaxInt, true
		}
		return math.MinInt, true
	default:
		return 0, false
	}
}

func floatToInt(x float64) int {
	if math.MinInt <= x && x <= math.MaxInt {
		return int(x)
	}
	if x > 0 {
		return math.MaxInt
	}
	return math.MinInt
}

func toFloat(x any) (float64, bool) {
	switch x := x.(type) {
	case int:
		return float64(x), true
	case float64:
		return x, true
	case *big.Int:
		return bigToFloat(x), true
	default:
		return 0.0, false
	}
}

func bigToFloat(x *big.Int) float64 {
	if x.IsInt64() {
		return float64(x.Int64())
	}
	if f, err := strconv.ParseFloat(x.String(), 64); err == nil {
		return f
	}
	return math.Inf(x.Sign())
}
