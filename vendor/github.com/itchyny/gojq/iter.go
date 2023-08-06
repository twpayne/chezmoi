package gojq

// Iter is an interface for an iterator.
type Iter interface {
	Next() (any, bool)
}

// NewIter creates a new [Iter] from values.
func NewIter(values ...any) Iter {
	switch len(values) {
	case 0:
		return emptyIter{}
	case 1:
		return &unitIter{value: values[0]}
	default:
		iter := sliceIter(values)
		return &iter
	}
}

type emptyIter struct{}

func (emptyIter) Next() (any, bool) {
	return nil, false
}

type unitIter struct {
	value any
	done  bool
}

func (iter *unitIter) Next() (any, bool) {
	if iter.done {
		return nil, false
	}
	iter.done = true
	return iter.value, true
}

type sliceIter []any

func (iter *sliceIter) Next() (any, bool) {
	if len(*iter) == 0 {
		return nil, false
	}
	value := (*iter)[0]
	*iter = (*iter)[1:]
	return value, true
}
