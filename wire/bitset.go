package wire

import (
	"fmt"

	"terorie.dev/nimiq/beserial"
)

type BitSet struct {
	Len  uint8
	Bits []byte
}

func (bs *BitSet) UnmarshalBESerial(b []byte) (n int, err error) {
	// Read count
	if len(b) < 1 {
		return 0, fmt.Errorf("reading count: %w", beserial.ErrUnexpectedEOF)
	}
	bs.Len = b[0]
	b = b[1:]
	// Read repeat bits
	size := (bs.Len + 7) / 8 // ceil division
	bs.Bits = make([]byte, size)
	if len(b) < len(bs.Bits) {
		return 0, fmt.Errorf("reading bit set: %w", beserial.ErrUnexpectedEOF)
	}
	copy(bs.Bits, b)
	return 1 + int(size), nil
}

func (bs *BitSet) MarshalBESerial(b []byte) ([]byte, error) {
	b = append(b, bs.Len)
	b = append(b, bs.Bits...)
	return b, nil
}

func (bs *BitSet) SizeBESerial() (n int, err error) {
	return 1 + len(bs.Bits), nil
}

func (bs *BitSet) bit(i uint8) bool {
	return bs.Bits[i/8]&(0x80>>(i%8)) != 0
}
