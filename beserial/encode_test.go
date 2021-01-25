package beserial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	t.Run("Uint8", func(t *testing.T) {
		y := uint8(0xCD)
		data := []byte{0xCD}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("Struct", func(t *testing.T) {
		type s struct {
			U64 uint64
			U32 uint32
			U16 uint16
			U8  uint8
			I64 int64
			I32 int32
			I16 int16
			I8  int8
		}
		y := s{
			U64: 0xC0C0_4040_0909_0303,
			U32: 0x7070_D5D5,
			U16: 0x2349,
			U8:  0x20,
			I64: -0x2222_3333_4444_5555,
			I32: -0x4FEF_EFEF,
			I16: -2020,
			I8:  -128,
		}
		data := []byte{
			0xC0, 0xC0, 0x40, 0x40, 0x09, 0x09, 0x03, 0x03,
			0x70, 0x70, 0xD5, 0xD5,
			0x23, 0x49,
			0x20,
			0xDD, 0xDD, 0xCC, 0xCC, 0xBB, 0xBB, 0xAA, 0xAB,
			0xB0, 0x10, 0x10, 0x11,
			0xF8, 0x1C,
			0x80,
		}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("ByteSlice", func(t *testing.T) {
		var s struct {
			Data []byte `beserial:"len_tag=uint16"`
		}
		s.Data = []byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		}
		data := []byte{
			0x00, 16,
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		}
		b, err := Marshal(nil, &s)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("Slice", func(t *testing.T) {
		var s struct {
			Data []uint16 `beserial:"len_tag=uint16"`
		}
		s.Data = []uint16{0x1234, 0x5678}
		data := []byte{
			0x00, 2,
			0x12, 0x34, 0x56, 0x78,
		}
		b, err := Marshal(nil, &s)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("ByteArray", func(t *testing.T) {
		y := [16]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, y[:], b)
	})
	t.Run("Array", func(t *testing.T) {
		y := [2]uint16{0x1234, 0x5678}
		data := []byte{0x12, 0x34, 0x56, 0x78}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("String", func(t *testing.T) {
		var y struct {
			A string `beserial:"len_tag=uint8"`
		}
		y.A = "ya yeet"
		data := []byte{0x07, 0x79, 0x61, 0x20, 0x79, 0x65, 0x65, 0x74}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("Optional_None", func(t *testing.T) {
		type s struct {
			A uint8
			B *uint8 `beserial:"optional"`
			C uint8
		}
		y := s{
			A: 0x10,
			B: nil,
			C: 0x30,
		}
		data := []byte{0x10, 0x00, 0x30}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("Optional_Some", func(t *testing.T) {
		type s struct {
			A uint8
			B *uint8 `beserial:"optional"`
			C uint8
		}
		y := s{
			A: 0x10,
			B: new(uint8),
			C: 0x30,
		}
		data := []byte{0x10, 0x01, 0x00, 0x30}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
	t.Run("Custom", func(t *testing.T) {
		type s struct {
			A uint8
			B customMarshalTest
			C uint8
		}
		y := s{
			A: 0x11,
			B: customMarshalTest{},
			C: 0x22,
		}
		data := []byte{0x11, 0x99, 0xAA, 0x22}
		b, err := Marshal(nil, &y)
		assert.NoError(t, err)
		assert.Equal(t, data, b)
	})
}

type customMarshalTest struct{}

func (_ *customMarshalTest) MarshalBESerial(b []byte) ([]byte, error) {
	return append(b, 0x99, 0xAA), nil
}

func (_ *customMarshalTest) SizeBESerial() (int, error) {
	return 2, nil
}
