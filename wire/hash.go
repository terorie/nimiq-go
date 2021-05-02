package wire

import (
	"terorie.dev/nimiq/beserial"
)

// Hash describes one of the supported on-chain hashes.
type Hash struct {
	Algorithm uint8
	Bytes     [32]byte
}

// Supported hash types.
const (
	HashBlake2b = 1
	HashSha256  = 3
)

func (h *Hash) UnmarshalBESerial(b []byte) (n int, err error) {
	if len(b) < 1 {
		return 0, beserial.ErrUnexpectedEOF
	}
	algo := b[0]
	n++
	b = b[1:]
	if len(b) < 32 {
		return 0, beserial.ErrUnexpectedEOF
	}
	h.Algorithm = algo
	copy(h.Bytes[:], b)
	n += 32
	return
}

func (h *Hash) MarshalBESerial(b []byte) ([]byte, error) {
	b = append(b, h.Algorithm)
	b = append(b, h.Bytes[:]...)
	return b, nil
}

func (h *Hash) SizeBESerial() (int, error) {
	return 1 + len(h.Bytes), nil
}
