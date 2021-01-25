package beserial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSize(t *testing.T) {
	t.Run("Uint8", func(t *testing.T) {
		var x uint8
		z, err := Size(&x)
		assert.NoError(t, err)
		assert.Equal(t, 1, z)
	})
	t.Run("Struct", func(t *testing.T) {
		var x struct {
			U64 uint64
			U32 uint32
			U16 uint16
			U8  uint8
			I64 int64
			I32 int32
			I16 int16
			I8  int8
		}
		z, err := Size(&x)
		assert.NoError(t, err)
		assert.Equal(t, 30, z)
	})
	t.Run("NumberSlice", func(t *testing.T) {
		var s struct {
			Data []uint16 `beserial:"len_tag=uint16"`
		}
		s.Data = make([]uint16, 16)
		z, err := Size(&s)
		assert.NoError(t, err)
		assert.Equal(t, 34, z)
	})
	t.Run("Slice", func(t *testing.T) {
		type s struct {
			A, B, C uint8
		}
		var s2 struct {
			Data []s `beserial:"len_tag=uint16"`
		}
		s2.Data = make([]s, 16)
		z, err := Size(&s2)
		assert.NoError(t, err)
		assert.Equal(t, 50, z)
	})
	t.Run("NumberArray", func(t *testing.T) {
		var s [16]uint16
		z, err := Size(&s)
		assert.NoError(t, err)
		assert.Equal(t, 32, z)
	})
	t.Run("Array", func(t *testing.T) {
		type s struct {
			A, B, C uint8
		}
		var s2 struct {
			Data [16]s
		}
		z, err := Size(&s2)
		assert.NoError(t, err)
		assert.Equal(t, 48, z)
	})
	t.Run("String", func(t *testing.T) {
		var s struct {
			S string `beserial:"len_tag=uint8"`
		}
		s.S = "ya yeet"
		z, err := Size(&s)
		assert.NoError(t, err)
		assert.Equal(t, 8, z)
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
		z, err := Size(&y)
		assert.NoError(t, err)
		assert.Equal(t, 3, z)
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
		z, err := Size(&y)
		assert.NoError(t, err)
		assert.Equal(t, 4, z)
	})
}
