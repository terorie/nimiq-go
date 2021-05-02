package beserial

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

// Unmarshal parses the beserial-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
//
// The number of bytes read are stored in n,
// which is guaranteed to be 0 <= n <= len(data).
func Unmarshal(data []byte, v interface{}) (n int, err error) {
	d := decodeState{data: data}
	err = d.unmarshal(v)
	n = len(data) - len(d.data)
	return
}

// UnmarshalFull is like Unmarshal but errors if some bytes were not consumed.
func UnmarshalFull(data []byte, v interface{}) error {
	n, err := Unmarshal(data, v)
	if err != nil {
		return err
	}
	if len(data) > n {
		return fmt.Errorf("%d bytes not deserialized", len(data)-n)
	}
	return nil
}

// Unmarshaler is the interface implemented by types
// that can unmarshal a BESerial encoding of themselves.
type Unmarshaler interface {
	// UnmarshalBESerial takes some encoded input,
	// decodes it to itself and returns the number of bytes read.
	// All bytes afterwards are assumed to belong to a different item.
	UnmarshalBESerial([]byte) (int, error)
}

type decodeState struct {
	data []byte
}

func (d *decodeState) unmarshal(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return nil
	} else if rv.Kind() != reflect.Ptr {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	if err := d.value(rv.Elem(), tags{}); err != nil {
		return err
	}
	return nil
}

func (d *decodeState) value(v reflect.Value, ts tags) error {
	kind := v.Kind()
	// Check for UnmarshalBESerial implementation.
	var addr reflect.Value
	if kind == reflect.Ptr {
		addr = v
	} else if v.CanAddr() {
		addr = v.Addr()
	}
	if addr.IsValid() && addr.Type().NumMethod() > 0 && addr.CanInterface() {
		if u, ok := addr.Interface().(Unmarshaler); ok {
			n, err := u.UnmarshalBESerial(d.data)
			if err != nil {
				return err
			} else if n < 0 {
				panic("beserial: UnmarshalBESerial returned negative number")
			} else if n > len(d.data) {
				panic("beserial: UnmarshalBESerial claimed to have read more bytes than available")
			}
			d.data = d.data[n:]
			return nil
		}
	}
	// Generic decode.
	switch kind {
	case reflect.Ptr:
		typ := v.Type()
		if ts.optional {
			// If the optional struct tag is set,
			// and the optional flag == 0x00, break.
			flag, err := d.pop(1)
			if err != nil {
				return err
			}
			switch flag[0] {
			case 0x00:
				v.Set(reflect.Zero(typ))
				return nil
			case 0x01:
				break
			default:
				return fmt.Errorf("invalid optional flag: 0x%x", flag[0])
			}
		}
		// Allocate pointer and recurse to value.
		ptrV := reflect.New(typ.Elem())
		v.Set(ptrV)
		return d.value(ptrV.Elem(), ts)
	case reflect.Array:
		typ := v.Type()
		if typ.Elem().Kind() == reflect.Uint8 {
			// Byte arrays can be copied.
			b, err := d.pop(typ.Len())
			if err != nil {
				return err
			}
			reflect.Copy(v, reflect.ValueOf(b))
		} else {
			for i := 0; i < typ.Len(); i++ {
				if err := d.value(v.Index(i), tags{}); err != nil {
					return err
				}
			}
		}
	case reflect.Slice, reflect.String:
		if ts.lenTag == reflect.Invalid {
			return fmt.Errorf("in %s: %w", v.Type().String(), ErrNoLenTag)
		}
		typ := v.Type()
		u, i, signed, err := d.number(ts.lenTag)
		if err != nil {
			return err
		}
		if signed {
			if i < 0 {
				return fmt.Errorf("got slice with negative size")
			}
			u = uint64(i)
		}
		const maxInt = (^uint(0)) >> 1
		if u > uint64(maxInt) {
			return fmt.Errorf("got invalid slice len: %d", u)
		}
		if kind == reflect.String {
			// String contents can be copied.
			b, err := d.pop(int(u))
			if err != nil {
				return err
			}
			v.SetString(string(b))
		} else if typ.Elem().Kind() == reflect.Uint8 {
			// Byte slices can be copied.
			b, err := d.pop(int(u))
			if err != nil {
				return err
			}
			slice := reflect.MakeSlice(typ, int(u), int(u))
			copy(slice.Bytes(), b)
			v.Set(slice)
		} else {
			if u > uint64(len(d.data)) {
				// Sanity check: Defend against large, impossible allocations.
				// FIXME: Set a hard limit on alloc size
				return fmt.Errorf("cannot allocate slice of %d for only %d bytes",
					u, len(d.data))
			}
			v.Set(reflect.MakeSlice(typ, int(u), int(u)))
			for i := 0; i < int(u); i++ {
				if err := d.value(v.Index(i), tags{}); err != nil {
					return err
				}
			}
		}
	case reflect.Struct:
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			fieldType := typ.Field(i)
			fieldValue := v.Field(i)
			var elTs tags
			tag, ok := fieldType.Tag.Lookup("beserial")
			if ok {
				elTs.parse(tag)
			}
			if err := d.value(fieldValue, elTs); err != nil {
				return fmt.Errorf(`in "%s" field "%s": %w`,
					typ.String(), fieldType.Name, err)
			}
		}
	default:
		u, i, signed, err := d.number(kind)
		if err != nil {
			return err
		}
		if signed {
			v.SetInt(i)
		} else {
			v.SetUint(u)
		}
	}
	return nil
}

func (d *decodeState) number(kind reflect.Kind) (u uint64, i int64, signed bool, err error) {
	switch kind {
	case reflect.Uint8:
		b, err := d.pop(1)
		if err != nil {
			return 0, 0, false, err
		}
		return uint64(b[0]), 0, false, nil
	case reflect.Uint16:
		b, err := d.pop(2)
		if err != nil {
			return 0, 0, false, err
		}
		return uint64(binary.BigEndian.Uint16(b)), 0, false, nil
	case reflect.Uint32:
		b, err := d.pop(4)
		if err != nil {
			return 0, 0, false, err
		}
		return uint64(binary.BigEndian.Uint32(b)), 0, false, nil
	case reflect.Uint64:
		b, err := d.pop(8)
		if err != nil {
			return 0, 0, false, err
		}
		return binary.BigEndian.Uint64(b), 0, false, nil
	case reflect.Int8:
		b, err := d.pop(1)
		if err != nil {
			return 0, 0, false, err
		}
		return 0, int64(int8(b[0])), true, nil
	case reflect.Int16:
		b, err := d.pop(2)
		if err != nil {
			return 0, 0, false, err
		}
		return 0, int64(int16(binary.BigEndian.Uint16(b))), true, nil
	case reflect.Int32:
		b, err := d.pop(4)
		if err != nil {
			return 0, 0, false, err
		}
		return 0, int64(binary.BigEndian.Uint32(b)), true, nil
	case reflect.Int64:
		b, err := d.pop(8)
		if err != nil {
			return 0, 0, false, err
		}
		return 0, int64(binary.BigEndian.Uint64(b)), true, nil
	default:
		return 0, 0, false, fmt.Errorf("%s cannot be unmarshalled", kind)
	}
}

// pop returns the first n data bytes and removes them from the state.
func (d *decodeState) pop(n int) ([]byte, error) {
	if n > len(d.data) {
		return nil, ErrUnexpectedEOF
	}
	buf := d.data[:n]
	d.data = d.data[n:]
	return buf, nil
}
