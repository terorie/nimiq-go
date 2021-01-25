package address

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	correctAddr := [20]byte{
		0x21, 0xa9, 0x34, 0xfe, 0x3d,
		0x6a, 0x68, 0xbd, 0xb6, 0x44,
		0x47, 0xc5, 0x71, 0xc8, 0x8c,
		0x19, 0xe3, 0x9f, 0xb6, 0x85,
	}
	t.Run("NoSpaces", func(t *testing.T) {
		addr, err := Decode("NQ1946LK9YHVD9LBTDJ48Y2P3J4C37HRYDL5")
		assert.NoError(t, err)
		assert.Equal(t, correctAddr, addr)
	})
	t.Run("Normal", func(t *testing.T) {
		addr, err := Decode("NQ19 46LK 9YHV D9LB TDJ4 8Y2P 3J4C 37HR YDL5")
		assert.NoError(t, err)
		assert.Equal(t, correctAddr, addr)
	})
	t.Run("ErrChecksum", func(t *testing.T) {
		_, err := Decode("NQ20 46LK 9YHV D9LB TDJ4 8Y2P 3J4C 37HR YDL5")
		assert.Equal(t, ErrChecksum, err)
	})
	t.Run("ErrInvalidCountryCode", func(t *testing.T) {
		_, err := Decode("LD20 46LK 9YHV D9LB TDJ4 8Y2P 3J4C 37HR YDL5")
		assert.Equal(t, ErrInvalidCountryCode, err)
	})
	t.Run("ErrInvalidCharacter", func(t *testing.T) {
		_, err := Decode("NQ20 46LK 9YHV D9LB .... 8Y2P 3J4C 37HR YDL5")
		assert.Equal(t, ErrInvalidCharacter, err)
	})
	t.Run("TooShort", func(t *testing.T) {
		_, err := Decode("NQ20 46LK 9YHV D9LB IIII 8Y2P 3J4C 37HR YDL")
		assert.Equal(t, ErrInvalidLength, err)
	})
	t.Run("TooLong", func(t *testing.T) {
		_, err := Decode("NQ20 46LK 9YHV D9LB IIII 8Y2P 3J4C 37HR YDL5 A")
		assert.Equal(t, ErrInvalidLength, err)
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := Decode("                                            ")
		assert.Equal(t, ErrInvalidLength, err)
	})
	t.Run("CyrillicCharacter", func(t *testing.T) {
		_, err := Decode("NQ20 ААB 9YHV D9LB IIII 8Y2P 3J4C 37HR YDL")
		assert.Equal(t, ErrInvalidCharacter, err)
	})
}

func TestEncode(t *testing.T) {
	addrInput := [20]byte{
		0x21, 0xa9, 0x34, 0xfe, 0x3d,
		0x6a, 0x68, 0xbd, 0xb6, 0x44,
		0x47, 0xc5, 0x71, 0xc8, 0x8c,
		0x19, 0xe3, 0x9f, 0xb6, 0x85,
	}
	correctOutput := "NQ19 46LK 9YHV D9LB TDJ4 8Y2P 3J4C 37HR YDL5"

	// 1. Test if conversion works
	str1 := Encode(&addrInput)
	if str1 != correctOutput {
		t.Errorf("Invalid user friendly address encoded: %s\n", str1)
	}
}

func BenchmarkEncode(b *testing.B) {
	addrs := make([][20]byte, b.N)
	for i := 0; i < b.N; i++ {
		rand.Read(addrs[i][:])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if Encode(&addrs[i]) == "" {
			b.Fatal("Got empty address")
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	ibans := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		var addr [20]byte
		rand.Read(addr[:])
		ibans[i] = Encode(&addr)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decode(ibans[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}
