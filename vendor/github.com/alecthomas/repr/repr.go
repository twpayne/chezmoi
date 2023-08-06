// Package repr attempts to represent Go values in a form that can be copy-and-pasted into source
// code directly.
//
// Some values (such as pointers to basic types) can not be represented directly in
// Go. These values will be output as `&<value>`. eg. `&23`
package repr

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
	"unsafe"
)

var (
	// "Real" names of basic kinds, used to differentiate type aliases.
	realKindName = map[reflect.Kind]string{
		reflect.Bool:       "bool",
		reflect.Int:        "int",
		reflect.Int8:       "int8",
		reflect.Int16:      "int16",
		reflect.Int32:      "int32",
		reflect.Int64:      "int64",
		reflect.Uint:       "uint",
		reflect.Uint8:      "uint8",
		reflect.Uint16:     "uint16",
		reflect.Uint32:     "uint32",
		reflect.Uint64:     "uint64",
		reflect.Uintptr:    "uintptr",
		reflect.Float32:    "float32",
		reflect.Float64:    "float64",
		reflect.Complex64:  "complex64",
		reflect.Complex128: "complex128",
		reflect.Array:      "array",
		reflect.Chan:       "chan",
		reflect.Func:       "func",
		reflect.Map:        "map",
		reflect.Slice:      "slice",
		reflect.String:     "string",
	}

	goStringerType = reflect.TypeOf((*fmt.GoStringer)(nil)).Elem()
	anyType        = reflect.TypeOf((*any)(nil)).Elem()

	byteSliceType = reflect.TypeOf([]byte{})
)

// Default prints to os.Stdout with two space indentation.
var Default = New(os.Stdout, Indent("  "))

// An Option modifies the default behaviour of a Printer.
type Option func(o *Printer)

// Indent output by this much.
func Indent(indent string) Option { return func(o *Printer) { o.indent = indent } }

// NoIndent disables indenting.
func NoIndent() Option { return Indent("") }

// OmitEmpty sets whether empty field members should be omitted from output.
func OmitEmpty(omitEmpty bool) Option { return func(o *Printer) { o.omitEmpty = omitEmpty } }

// ExplicitTypes adds explicit typing to slice and map struct values that would normally be inferred by Go.
func ExplicitTypes(ok bool) Option { return func(o *Printer) { o.explicitTypes = true } }

// IgnoreGoStringer disables use of the .GoString() method.
func IgnoreGoStringer() Option { return func(o *Printer) { o.ignoreGoStringer = true } }

// Hide excludes the given types from representation, instead just printing the name of the type.
func Hide(ts ...any) Option {
	return func(o *Printer) {
		for _, t := range ts {
			rt := reflect.Indirect(reflect.ValueOf(t)).Type()
			o.exclude[rt] = true
		}
	}
}

// AlwaysIncludeType always includes explicit type information for each item.
func AlwaysIncludeType() Option { return func(o *Printer) { o.alwaysIncludeType = true } }

// Printer represents structs in a printable manner.
type Printer struct {
	indent            string
	omitEmpty         bool
	ignoreGoStringer  bool
	alwaysIncludeType bool
	explicitTypes     bool
	exclude           map[reflect.Type]bool
	w                 io.Writer
}

// New creates a new Printer on w with the given Options.
func New(w io.Writer, options ...Option) *Printer {
	p := &Printer{
		w:         w,
		indent:    "  ",
		omitEmpty: true,
		exclude:   map[reflect.Type]bool{},
	}
	for _, option := range options {
		option(p)
	}
	return p
}

func (p *Printer) nextIndent(indent string) string {
	if p.indent != "" {
		return indent + p.indent
	}
	return ""
}

func (p *Printer) thisIndent(indent string) string {
	if p.indent != "" {
		return indent
	}
	return ""
}

// Print the values.
func (p *Printer) Print(vs ...any) {
	for i, v := range vs {
		if i > 0 {
			fmt.Fprint(p.w, " ")
		}
		p.reprValue(map[reflect.Value]bool{}, reflect.ValueOf(v), "", true, false)
	}
}

// Println prints each value on a new line.
func (p *Printer) Println(vs ...any) {
	for i, v := range vs {
		if i > 0 {
			fmt.Fprint(p.w, " ")
		}
		p.reprValue(map[reflect.Value]bool{}, reflect.ValueOf(v), "", true, false)
	}
	fmt.Fprintln(p.w)
}

// showType is true if struct types should be shown. isAnyValue is true if the containing value is an "any" type.
func (p *Printer) reprValue(seen map[reflect.Value]bool, v reflect.Value, indent string, showStructType bool, isAnyValue bool) { // nolint: gocyclo
	if seen[v] {
		fmt.Fprint(p.w, "...")
		return
	}
	seen[v] = true
	defer delete(seen, v)

	if v.Kind() == reflect.Invalid || (v.Kind() == reflect.Ptr || v.Kind() == reflect.Map || v.Kind() == reflect.Chan || v.Kind() == reflect.Slice || v.Kind() == reflect.Func || v.Kind() == reflect.Interface) && v.IsNil() {
		fmt.Fprint(p.w, "nil")
		return
	}
	if p.exclude[v.Type()] {
		fmt.Fprintf(p.w, "%s...", v.Type().Name())
		return
	}
	t := v.Type()

	if t == byteSliceType {
		fmt.Fprintf(p.w, "[]byte(%q)", v.Bytes())
		return
	}

	// If we can't access a private field directly with reflection, try and do so via unsafe.
	if !v.CanInterface() && v.CanAddr() {
		uv := reflect.NewAt(t, unsafe.Pointer(v.UnsafeAddr())).Elem()
		if uv.CanInterface() {
			v = uv
		}
	}
	// Attempt to use fmt.GoStringer interface.
	if !p.ignoreGoStringer && t.Implements(goStringerType) {
		fmt.Fprint(p.w, v.Interface().(fmt.GoStringer).GoString())
		return
	}
	in := p.thisIndent(indent)
	ni := p.nextIndent(indent)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		fmt.Fprintf(p.w, "%s{", substAny(v.Type()))
		if v.Len() == 0 {
			fmt.Fprint(p.w, "}")
		} else {
			if p.indent != "" {
				fmt.Fprintf(p.w, "\n")
			}
			for i := 0; i < v.Len(); i++ {
				e := v.Index(i)
				fmt.Fprintf(p.w, "%s", ni)
				p.reprValue(seen, e, ni, p.alwaysIncludeType || p.explicitTypes, v.Type().Elem() == anyType)
				if p.indent != "" {
					fmt.Fprintf(p.w, ",\n")
				} else if i < v.Len()-1 {
					fmt.Fprintf(p.w, ", ")
				}
			}
			fmt.Fprintf(p.w, "%s}", in)
		}

	case reflect.Chan:
		fmt.Fprintf(p.w, "make(")
		fmt.Fprintf(p.w, "%s", substAny(v.Type()))
		fmt.Fprintf(p.w, ", %d)", v.Cap())

	case reflect.Map:
		fmt.Fprintf(p.w, "%s{", substAny(v.Type()))
		if p.indent != "" && v.Len() != 0 {
			fmt.Fprintf(p.w, "\n")
		}
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
		})
		for i, k := range keys {
			kv := v.MapIndex(k)
			fmt.Fprintf(p.w, "%s", ni)
			p.reprValue(seen, k, ni, p.alwaysIncludeType || p.explicitTypes, v.Type().Key() == anyType)
			fmt.Fprintf(p.w, ": ")
			p.reprValue(seen, kv, ni, true, v.Type().Elem() == anyType)
			if p.indent != "" {
				fmt.Fprintf(p.w, ",\n")
			} else if i < v.Len()-1 {
				fmt.Fprintf(p.w, ", ")
			}
		}
		fmt.Fprintf(p.w, "%s}", in)

	case reflect.Struct:
		if td, ok := asTime(v); ok {
			timeToGo(p.w, td)
		} else {
			if showStructType {
				fmt.Fprintf(p.w, "%s{", substAny(v.Type()))
			} else {
				fmt.Fprint(p.w, "{")
			}
			if p.indent != "" && v.NumField() != 0 {
				fmt.Fprintf(p.w, "\n")
			}
			for i := 0; i < v.NumField(); i++ {
				t := v.Type().Field(i)
				f := v.Field(i)
				if p.omitEmpty && f.IsZero() {
					continue
				}
				fmt.Fprintf(p.w, "%s%s: ", ni, t.Name)
				p.reprValue(seen, f, ni, true, t.Type == anyType)
				if p.indent != "" {
					fmt.Fprintf(p.w, ",\n")
				} else if i < v.NumField()-1 {
					fmt.Fprintf(p.w, ", ")
				}
			}
			fmt.Fprintf(p.w, "%s}", indent)
		}
	case reflect.Ptr:
		if v.IsNil() {
			fmt.Fprintf(p.w, "nil")
			return
		}
		if showStructType {
			fmt.Fprintf(p.w, "&")
		}
		p.reprValue(seen, v.Elem(), indent, showStructType, false)

	case reflect.String:
		if t.Name() != "string" || p.alwaysIncludeType {
			fmt.Fprintf(p.w, "%s(%q)", t, v.String())
		} else {
			fmt.Fprintf(p.w, "%q", v.String())
		}

	case reflect.Interface:
		if v.IsNil() {
			fmt.Fprintf(p.w, "%s(nil)", substAny(v.Type()))
		} else {
			p.reprValue(seen, v.Elem(), indent, true, true)
		}

	case reflect.Func:
		fmt.Fprint(p.w, substAny(v.Type()))

	default:
		if t.Name() != realKindName[t.Kind()] || p.alwaysIncludeType || isAnyValue {
			fmt.Fprintf(p.w, "%s(%v)", t, v)
		} else {
			fmt.Fprintf(p.w, "%v", v)
		}
	}
}

func asTime(v reflect.Value) (time.Time, bool) {
	if !v.CanInterface() {
		return time.Time{}, false
	}
	t, ok := v.Interface().(time.Time)
	return t, ok
}

// String returns a string representing v.
func String(v any, options ...Option) string {
	w := bytes.NewBuffer(nil)
	options = append([]Option{NoIndent()}, options...)
	p := New(w, options...)
	p.Print(v)
	return w.String()
}

func extractOptions(vs ...any) (args []any, options []Option) {
	for _, v := range vs {
		if o, ok := v.(Option); ok {
			options = append(options, o)
		} else {
			args = append(args, v)
		}
	}
	return
}

// Println prints v to os.Stdout, one per line.
func Println(vs ...any) {
	args, options := extractOptions(vs...)
	New(os.Stdout, options...).Println(args...)
}

// Print writes a representation of v to os.Stdout, separated by spaces.
func Print(vs ...any) {
	args, options := extractOptions(vs...)
	New(os.Stdout, options...).Print(args...)
}

func timeToGo(w io.Writer, t time.Time) {
	if t.IsZero() {
		fmt.Fprint(w, "time.Time{}")
		return
	}

	var zone string
	switch loc := t.Location(); loc {
	case nil:
		zone = "nil"
	case time.UTC:
		zone = "time.UTC"
	case time.Local:
		zone = "time.Local"
	default:
		n, off := t.Zone()
		zone = fmt.Sprintf("time.FixedZone(%q, %d)", n, off)
	}
	y, m, d := t.Date()
	fmt.Fprintf(w, `time.Date(%d, %d, %d, %d, %d, %d, %d, %s)`, y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), zone)
}

// Replace "interface {}" with "any"
func substAny(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), substAny(t.Elem()))

	case reflect.Slice:
		return "[]" + substAny(t.Elem())

	case reflect.Map:
		return "map[" + substAny(t.Key()) + "]" + substAny(t.Elem())

	case reflect.Chan:
		return fmt.Sprintf("%s %s", t.ChanDir(), substAny(t.Elem()))

	case reflect.Func:
		in := []string{}
		out := []string{}
		for i := 0; i < t.NumIn(); i++ {
			in = append(in, substAny(t.In(i)))
		}
		for i := 0; i < t.NumOut(); i++ {
			out = append(out, substAny(t.Out(i)))
		}
		if len(out) == 0 {
			return "func" + t.Name() + "(" + strings.Join(in, ", ") + ")"
		}
		return "func" + t.Name() + "(" + strings.Join(in, ", ") + ") (" + strings.Join(out, ", ") + ")"
	}

	if t == anyType {
		return "any"
	}
	return t.String()
}
