package beserial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	t.Run("Uint8", func(t *testing.T) {
		var x uint8
		assert.NoError(t, UnmarshalFull([]byte{0xFE}, &x))
		assert.Equal(t, uint8(0xFE), x)
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
		x := s{
			U64: 0xC0C0_4040_0909_0303,
			U32: 0x7070_D5D5,
			U16: 0x2349,
			U8:  0x20,
			I64: -0x2222_3333_4444_5555,
			I32: -0x4FEF_EFEF,
			I16: -2020,
			I8:  -128,
		}
		var y s
		assert.NoError(t, UnmarshalFull([]byte{
			0xC0, 0xC0, 0x40, 0x40, 0x09, 0x09, 0x03, 0x03,
			0x70, 0x70, 0xD5, 0xD5,
			0x23, 0x49,
			0x20,
			0xDD, 0xDD, 0xCC, 0xCC, 0xBB, 0xBB, 0xAA, 0xAB,
			0xB0, 0x10, 0x10, 0x11,
			0xF8, 0x1C,
			0x80,
		}, &y))
		assert.Equal(t, &x, &y)
	})
	t.Run("ByteSlice", func(t *testing.T) {
		type s struct {
			Data []byte `beserial:"len_tag=uint16"`
		}
		data := []byte{
			0x00, 16,
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		}
		var x s
		assert.NoError(t, UnmarshalFull(data, &x))
		assert.Equal(t, data[2:], x.Data)
	})
	t.Run("Slice", func(t *testing.T) {
		type s struct {
			Data []uint16 `beserial:"len_tag=uint16"`
		}
		data := []byte{
			0x00, 2,
			0x12, 0x34, 0x56, 0x78,
		}
		correct := []uint16{0x1234, 0x5678}
		var x s
		assert.NoError(t, UnmarshalFull(data, &x))
		assert.Equal(t, &correct, &x.Data)
	})
	t.Run("ByteArray", func(t *testing.T) {
		var x [16]byte
		y := [16]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		}
		assert.NoError(t, UnmarshalFull(y[:], &x))
		assert.Equal(t, y, x)
	})
	t.Run("Array", func(t *testing.T) {
		var x [2]uint16
		y := [2]uint16{0x1234, 0x5678}
		data := []byte{0x12, 0x34, 0x56, 0x78}
		assert.NoError(t, UnmarshalFull(data, &x))
		assert.Equal(t, &y, &x)
	})
	t.Run("String", func(t *testing.T) {
		type s struct {
			S string `beserial:"len_tag=uint8"`
		}
		var x s
		y := s{S: "ya yeet"}
		data := []byte{0x07, 0x79, 0x61, 0x20, 0x79, 0x65, 0x65, 0x74}
		assert.NoError(t, UnmarshalFull(data, &x))
		assert.Equal(t, &y, &x)
	})
	t.Run("Optional_None", func(t *testing.T) {
		type s struct {
			A uint8
			B *uint8 `beserial:"optional"`
			C uint8
		}
		var x s
		y := s{
			A: 0x10,
			B: nil,
			C: 0x30,
		}
		data := []byte{0x10, 0x00, 0x30}
		assert.NoError(t, UnmarshalFull(data, &x))
		assert.Equal(t, &y, &x)
	})
	t.Run("Optional_Some", func(t *testing.T) {
		type s struct {
			A uint8
			B *uint8 `beserial:"optional"`
			C uint8
		}
		var x s
		y := s{
			A: 0x10,
			B: new(uint8),
			C: 0x30,
		}
		data := []byte{0x10, 0x01, 0x00, 0x30}
		assert.NoError(t, UnmarshalFull(data, &x))
		assert.Equal(t, &y, &x)
	})
	t.Run("Custom", func(t *testing.T) {
		var x, y customUnmarshalTest
		data := []byte{0x00, 0x00, 0x00, 0x00}
		assert.NoError(t, UnmarshalFull(data, &x))
		assert.Equal(t, &y, &y)
	})
}

type customUnmarshalTest struct{}

func (*customUnmarshalTest) UnmarshalBESerial(_ []byte) (int, error) {
	return 4, nil
}
