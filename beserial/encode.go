package beserial

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

// Marshal appends the binary encoding of v to b.
func Marshal(b []byte, v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Interface {
		if rv.Interface() == nil {
			return nil, fmt.Errorf("TODO error") // FIXME
		}
	} else {
		if rv.IsNil() {
			return nil, fmt.Errorf("TODO error") // FIXME
		} else if rv.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("TODO error") // FIXME
		}
	}
	return marshal(rv.Elem(), b, tags{})
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid beserial.
type Marshaler interface {
	MarshalBESerial([]byte) ([]byte, error)
	SizeBESerial() (int, error)
}

func marshal(v reflect.Value, b []byte, ts tags) ([]byte, error) {
	var err error
	kind := v.Kind()
	// Check for MarshalBESerial implementation.
	var addr reflect.Value
	if kind == reflect.Ptr {
		addr = v
	} else if v.CanAddr() {
		addr = v.Addr()
	}
	if addr.IsValid() && addr.Type().NumMethod() > 0 && addr.CanInterface() {
		if u, ok := addr.Interface().(Marshaler); ok {
			return u.MarshalBESerial(b)
		}
	}
	// Generic encode.
	switch kind {
	case reflect.Ptr:
		if ts.optional {
			// If the optional struct tag is set,
			// and the ptr is nil, break.
			if v.IsNil() {
				b = append(b, 0x00)
				return b, nil
			}
			b = append(b, 0x01)
		} else if v.IsNil() {
			return nil, ErrNonOptionalNil
		}
		b, err = marshal(v.Elem(), b, ts)
		return b, nil
	case reflect.Array:
		typ := v.Type()
		if typ.Elem().Kind() == reflect.Uint8 {
			// Byte arrays can be copied.
			b = append(b, v.Slice(0, typ.Len()).Bytes()...)
		} else {
			for i := 0; i < typ.Len(); i++ {
				b, err = marshal(v.Index(i), b, tags{})
				if err != nil {
					return nil, err
				}
			}
		}
	case reflect.Slice, reflect.String:
		if ts.lenTag == reflect.Invalid {
			return nil, ErrNoLenTag
		}
		size := v.Len()
		b = marshalInt(ts.lenTag, size, b)
		typ := v.Type()
		if kind == reflect.String {
			b = append(b, []byte(v.String())...)
		} else if typ.Elem().Kind() == reflect.Uint8 {
			b = append(b, v.Bytes()...)
		} else {
			for i := 0; i < size; i++ {
				b, err = marshal(v.Index(i), b, tags{})
				if err != nil {
					return nil, err
				}
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
			b, err = marshal(sv, b, elTs)
			if err != nil {
				return nil, err
			}
		}
	default:
		var ok bool
		b, ok = marshalNumber(kind, v, b)
		if !ok {
			return nil, fmt.Errorf("%s cannot be marshalled", kind)
		}
	}
	return b, nil
}

// TODO Bounds checks

func marshalInt(kind reflect.Kind, n int, b []byte) []byte {
	var st [8]byte
	switch kind {
	case reflect.Uint8:
		return append(b, uint8(n))
	case reflect.Uint16:
		binary.BigEndian.PutUint16(st[:], uint16(n))
		return append(b, st[:2]...)
	case reflect.Uint32:
		binary.BigEndian.PutUint32(st[:], uint32(n))
		return append(b, st[:4]...)
	case reflect.Uint64:
		binary.BigEndian.PutUint64(st[:], uint64(n))
		return append(b, st[:8]...)
	default:
		panic("beserial: marshalInt not expecting " + kind.String())
	}
}

func marshalNumber(kind reflect.Kind, v reflect.Value, b []byte) ([]byte, bool) {
	var st [8]byte
	switch kind {
	case reflect.Int8:
		return append(b, uint8(int8(v.Int()))), true
	case reflect.Int16:
		binary.BigEndian.PutUint16(st[:], uint16(int16(v.Int())))
		return append(b, st[:2]...), true
	case reflect.Int32:
		binary.BigEndian.PutUint32(st[:], uint32(int32(v.Int())))
		return append(b, st[:4]...), true
	case reflect.Int64:
		binary.BigEndian.PutUint64(st[:], uint64(v.Int()))
		return append(b, st[:8]...), true
	case reflect.Uint8:
		return append(b, uint8(v.Uint())), true
	case reflect.Uint16:
		binary.BigEndian.PutUint16(st[:], uint16(v.Uint()))
		return append(b, st[:2]...), true
	case reflect.Uint32:
		binary.BigEndian.PutUint32(st[:], uint32(v.Uint()))
		return append(b, st[:4]...), true
	case reflect.Uint64:
		binary.BigEndian.PutUint64(st[:], v.Uint())
		return append(b, st[:8]...), true
	default:
		return b, false
	}
}
