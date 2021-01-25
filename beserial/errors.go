package beserial

import (
	"errors"
	"reflect"
)

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "beserial: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Ptr {
		return "beserial: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "beserial: Unmarshal(nil " + e.Type.String() + ")"
}

// An InvalidLenTagError describes an invalid or missing "len_tag" in a struct.
type InvalidLenTagError struct {
	Type reflect.Type
}

func (e *InvalidLenTagError) Error() string {
	return "beserial: invalid slice length tag: " + e.Type.String()
}

// Error constants
var (
	ErrUnexpectedEOF  = errors.New("beserial: unexpected EOF")
	ErrNoLenTag       = errors.New("beserial: no length tag specified for slice")
	ErrNonOptionalNil = errors.New("beserial: nil ptr without optional tag")
)
