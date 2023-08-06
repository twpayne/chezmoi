// Copyright (c) 2018, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package expand

import (
	"runtime"
	"sort"
	"strings"
)

// Environ is the base interface for a shell's environment, allowing it to fetch
// variables by name and to iterate over all the currently set variables.
type Environ interface {
	// Get retrieves a variable by its name. To check if the variable is
	// set, use Variable.IsSet.
	Get(name string) Variable

	// Each iterates over all the currently set variables, calling the
	// supplied function on each variable. Iteration is stopped if the
	// function returns false.
	//
	// The names used in the calls aren't required to be unique or sorted.
	// If a variable name appears twice, the latest occurrence takes
	// priority.
	//
	// Each is required to forward exported variables when executing
	// programs.
	Each(func(name string, vr Variable) bool)
}

// WriteEnviron is an extension on Environ that supports modifying and deleting
// variables.
type WriteEnviron interface {
	Environ
	// Set sets a variable by name. If !vr.IsSet(), the variable is being
	// unset; otherwise, the variable is being replaced.
	//
	// It is the implementation's responsibility to handle variable
	// attributes correctly. For example, changing an exported variable's
	// value does not unexport it, and overwriting a name reference variable
	// should modify its target.
	//
	// An error may be returned if the operation is invalid, such as if the
	// name is empty or if we're trying to overwrite a read-only variable.
	Set(name string, vr Variable) error
}

type ValueKind uint8

const (
	Unset ValueKind = iota
	String
	NameRef
	Indexed
	Associative
)

// Variable describes a shell variable, which can have a number of attributes
// and a value.
//
// A Variable is unset if its Kind field is Unset, which can be checked via
// [Variable.IsSet]. The zero value of a Variable is thus a valid unset variable.
//
// If a variable is set, its Value field will be a []string if it is an indexed
// array, a map[string]string if it's an associative array, or a string
// otherwise.
type Variable struct {
	Local    bool
	Exported bool
	ReadOnly bool

	Kind ValueKind

	Str  string            // Used when Kind is String or NameRef.
	List []string          // Used when Kind is Indexed.
	Map  map[string]string // Used when Kind is Associative.
}

// IsSet returns whether the variable is set. An empty variable is set, but an
// undeclared variable is not.
func (v Variable) IsSet() bool {
	return v.Kind != Unset
}

// String returns the variable's value as a string. In general, this only makes
// sense if the variable has a string value or no value at all.
func (v Variable) String() string {
	switch v.Kind {
	case String:
		return v.Str
	case Indexed:
		if len(v.List) > 0 {
			return v.List[0]
		}
	case Associative:
		// nothing to do
	}
	return ""
}

// maxNameRefDepth defines the maximum number of times to follow references when
// resolving a variable. Otherwise, simple name reference loops could crash a
// program quite easily.
const maxNameRefDepth = 100

// Resolve follows a number of nameref variables, returning the last reference
// name that was followed and the variable that it points to.
func (v Variable) Resolve(env Environ) (string, Variable) {
	name := ""
	for i := 0; i < maxNameRefDepth; i++ {
		if v.Kind != NameRef {
			return name, v
		}
		name = v.Str // keep name for the next iteration
		v = env.Get(name)
	}
	return name, Variable{}
}

// FuncEnviron wraps a function mapping variable names to their string values,
// and implements Environ. Empty strings returned by the function will be
// treated as unset variables. All variables will be exported.
//
// Note that the returned Environ's Each method will be a no-op.
func FuncEnviron(fn func(string) string) Environ {
	return funcEnviron(fn)
}

type funcEnviron func(string) string

func (f funcEnviron) Get(name string) Variable {
	value := f(name)
	if value == "" {
		return Variable{}
	}
	return Variable{Exported: true, Kind: String, Str: value}
}

func (f funcEnviron) Each(func(name string, vr Variable) bool) {}

// ListEnviron returns an Environ with the supplied variables, in the form
// "key=value". All variables will be exported. The last value in pairs is used
// if multiple values are present.
//
// On Windows, where environment variable names are case-insensitive, the
// resulting variable names will all be uppercase.
func ListEnviron(pairs ...string) Environ {
	return listEnvironWithUpper(runtime.GOOS == "windows", pairs...)
}

// listEnvironWithUpper implements ListEnviron, but letting the tests specify
// whether to uppercase all names or not.
func listEnvironWithUpper(upper bool, pairs ...string) Environ {
	list := append([]string{}, pairs...)
	if upper {
		// Uppercase before sorting, so that we can remove duplicates
		// without the need for linear search nor a map.
		for i, s := range list {
			if sep := strings.IndexByte(s, '='); sep > 0 {
				list[i] = strings.ToUpper(s[:sep]) + s[sep:]
			}
		}
	}

	sort.SliceStable(list, func(i, j int) bool {
		isep := strings.IndexByte(list[i], '=')
		jsep := strings.IndexByte(list[j], '=')
		if isep < 0 {
			isep = 0
		} else {
			isep += 1
		}
		if jsep < 0 {
			jsep = 0
		} else {
			jsep += 1
		}
		return list[i][:isep] < list[j][:jsep]
	})

	last := ""
	for i := 0; i < len(list); {
		s := list[i]
		sep := strings.IndexByte(s, '=')
		if sep <= 0 {
			// invalid element; remove it
			list = append(list[:i], list[i+1:]...)
			continue
		}
		name := s[:sep]
		if last == name {
			// duplicate; the last one wins
			list = append(list[:i-1], list[i:]...)
			continue
		}
		last = name
		i++
	}
	return listEnviron(list)
}

// listEnviron is a sorted list of "name=value" strings.
type listEnviron []string

func (l listEnviron) Get(name string) Variable {
	prefix := name + "="
	i := sort.SearchStrings(l, prefix)
	if i < len(l) && strings.HasPrefix(l[i], prefix) {
		return Variable{Exported: true, Kind: String, Str: strings.TrimPrefix(l[i], prefix)}
	}
	return Variable{}
}

func (l listEnviron) Each(fn func(name string, vr Variable) bool) {
	for _, pair := range l {
		i := strings.IndexByte(pair, '=')
		if i < 0 {
			// should never happen; see listEnvironWithUpper
			panic("expand.listEnviron: did not expect malformed name-value pair: " + pair)
		}
		name, value := pair[:i], pair[i+1:]
		if !fn(name, Variable{Exported: true, Kind: String, Str: value}) {
			return
		}
	}
}
