package wire

import (
	"fmt"

	"terorie.dev/nimiq/beserial"
)

// Hash describes one of the supported on-chain hashes.
type Hash struct {
	Algorithm uint8
	Bytes     []byte
}

// Supported hash types.
const (
	HashInvalid = uint8(iota)
	HashBlake2b = 1
	HashArgon2d = 2
	HashSha256  = 3
	HashSha512  = 4
)

// Sizes of hash algorithms.
const (
	HashBlake2bSize = 32
	HashArgon2dSize = 32
	HashSha256Size  = 32
	HashSha512Size  = 64
)

func (h *Hash) UnmarshalBESerial(b []byte) (n int, err error) {
	if len(b) < 1 {
		return 0, beserial.ErrUnexpectedEOF
	}
	algo := b[0]
	n++
	b = b[1:]
	var size int
	switch algo {
	case HashBlake2b:
		size = HashBlake2bSize
	case HashArgon2d:
		size = HashArgon2dSize
	case HashSha256:
		size = HashSha256Size
	case HashSha512:
		size = HashSha512Size
	default:
		return 0, fmt.Errorf("invalid hash type: 0x%x", h.Algorithm)
	}
	if len(b) < size {
		return 0, beserial.ErrUnexpectedEOF
	}
	h.Bytes = make([]byte, size)
	copy(h.Bytes, b)
	n += size
	return
}

func (h *Hash) MarshalBESerial(b []byte) ([]byte, error) {
	b = append(b, h.Algorithm)
	b = append(b, h.Bytes...)
	return b, nil
}

func (h *Hash) SizeBESerial() (int, error) {
	return 1 + len(h.Bytes), nil
}
