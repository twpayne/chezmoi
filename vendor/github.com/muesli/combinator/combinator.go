package combinator

import (
	"fmt"
	"reflect"
)

// Generate returns a slice of all possible value combinations for any given
// struct and a set of its potential member values.
func Generate(v interface{}, ov interface{}) error {
	// verify supplied types
	vPtr := reflect.ValueOf(v)
	if vPtr.Kind() != reflect.Ptr {
		return fmt.Errorf("non-pointer type supplied")
	}

	value := vPtr.Elem()
	if value.Kind() != reflect.Slice {
		return fmt.Errorf("pointer to non-slice type supplied")
	}
	vType := reflect.TypeOf(v).Elem().Elem()

	ovType := reflect.TypeOf(ov)
	ovValue := reflect.ValueOf(ov)
	if ovValue.Kind() != reflect.Struct {
		return fmt.Errorf("non-slice type supplied")
	}

	// calculate combinations
	combinations := 1
	members := ovType.NumField()
	for i := 0; i < members; i++ {
		if ovValue.Field(i).Kind() != reflect.Slice {
			continue
		}
		if ovValue.Field(i).Len() == 0 {
			// ignore empty option values
			continue
		}

		fname := ovType.Field(i).Name
		if _, ok := vType.FieldByName(fname); !ok {
			return fmt.Errorf("can't access struct field %s", fname)
		}

		combinations *= ovValue.Field(i).Len()
	}

	// fill struct with all combinations
	for i := 0; i < combinations; i++ {
		vi := reflect.Indirect(reflect.New(vType))

		offset := 1
		for j := 0; j < members; j++ {
			ovf := ovValue.Field(j)
			var fvalue reflect.Value

			if ovf.Kind() == reflect.Slice {
				if ovf.Len() == 0 {
					// ignore empty option values
					continue
				}

				fvalue = ovf.Index((i / offset) % ovf.Len())
				offset *= ovf.Len()
			} else {
				fvalue = ovf
			}

			fname := ovType.Field(j).Name
			if vi.FieldByName(fname).CanSet() {
				vi.FieldByName(fname).Set(fvalue)
			}

			// fmt.Println(fname, fvalue, offset)
		}

		// append item to original slice
		value.Set(reflect.Append(value, vi))
	}

	return nil
}
