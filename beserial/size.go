package beserial

import (
	"fmt"
	"reflect"
)

// Size returns the exact size of the result of
// beserial-marshalling the value pointed to by v.
// v must not be nil or a non-pointer.
func Size(v interface{}) (int, error) {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return 0, nil
	} else if rv.Kind() != reflect.Ptr {
		return 0, &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	return valueSize(rv.Elem(), tags{})
}

func valueSize(v reflect.Value, ts tags) (total int, err error) {
	kind := v.Kind()
	switch kind {
	case reflect.Ptr:
		if ts.optional {
			// If the optional struct tag is set,
			// and the ptr is nil, break.
			total++
			if v.IsNil() {
				return
			}
		} else if v.IsNil() {
			return 0, ErrNonOptionalNil
		}
		elem, err := valueSize(v.Elem(), ts)
		if err != nil {
			return 0, err
		}
		total += elem
	case reflect.Array:
		typ := v.Type()
		if elemSize, ok := sizeNumber(typ.Elem().Kind()); ok {
			total = typ.Len() * elemSize
		} else {
			for i := 0; i < typ.Len(); i++ {
				el, err := valueSize(v.Index(i), tags{})
				if err != nil {
					return 0, err
				}
				total += el
			}
		}
	case reflect.Slice, reflect.String:
		typ := v.Type()
		tagLen, ok := sizeNumber(ts.lenTag)
		if !ok {
			return 0, fmt.Errorf("slice without len tag")
		}
		total += tagLen
		if kind == reflect.String {
			total += v.Len()
		} else if elemSize, ok := sizeNumber(typ.Elem().Kind()); ok {
			total += v.Len() * elemSize
		} else {
			for i := 0; i < v.Len(); i++ {
				el, err := valueSize(v.Index(i), tags{})
				if err != nil {
					return 0, err
				}
				total += el
			}
		}
	case reflect.Struct:
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			var elTs tags
			tag, ok := typ.Field(i).Tag.Lookup("beserial")
			if ok {
				elTs.parse(tag)
			}
			sv := v.Field(i)
			elemSize, err := valueSize(sv, elTs)
			if err != nil {
				return 0, err
			}
			total += elemSize
		}
	default:
		tagSize, ok := sizeNumber(kind)
		if !ok {
			return 0, fmt.Errorf("%s cannot be marshalled", kind)
		}
		total += tagSize
	}
	return
}

func sizeNumber(kind reflect.Kind) (n int, ok bool) {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		return 1, true
	case reflect.Int16, reflect.Uint16:
		return 2, true
	case reflect.Int32, reflect.Uint32:
		return 4, true
	case reflect.Int64, reflect.Uint64:
		return 8, true
	default:
		return 0, false
	}
}
