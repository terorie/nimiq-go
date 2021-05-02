package wire

import (
	"fmt"
	"math/bits"

	"terorie.dev/nimiq/beserial"
)

// A Block in the block chain!
type Block struct {
	Header    BlockHeader
	Interlink BlockInterlink
	Body      *BlockBody `beserial:"optional"`
}

// BlockHeader identifies a block and defines various bits of chain state.
type BlockHeader struct {
	Version       uint16
	PrevHash      [32]byte
	InterlinkHash [32]byte
	BodyHash      [32]byte
	AccountsHash  [32]byte
	NBits         uint32
	Height        uint32
	Timestamp     uint32
	Nonce         uint32
}

// BlockBody holds the transactions in a block.
type BlockBody struct {
	MinerAddr [20]byte
	ExtraData []byte          `beserial:"len_tag=uint8"`
	Txs       []WrapTx        `beserial:"len_tag=uint16"`
	Pruned    []AccountPruned `beserial:"len_tag=uint16"`
}

// BlockInterlink builds the NIPoPoW proofs.
type BlockInterlink struct {
	Hashes     []*[32]byte // pointers to Compressed
	PrevHash   [32]byte
	RepeatBits []byte
	Compressed [][32]byte
}

// UnmarshalBESerial implements the interlink decoding algorithm.
func (il *BlockInterlink) UnmarshalBESerial(b []byte) (n int, err error) {
	orig := b
	// Read count
	if len(b) < 1 {
		return 0, fmt.Errorf("failed to read count: %w", beserial.ErrUnexpectedEOF)
	}
	count := b[0]
	b = b[1:]
	il.Hashes = make([]*[32]byte, count)
	// Read repeat bits
	repeatBitsSize := (count + 7) / 8 // ceil division
	il.RepeatBits = make([]byte, repeatBitsSize)
	if len(b) < len(il.RepeatBits) {
		return 0, fmt.Errorf("failed to read repeat bits: %w", beserial.ErrUnexpectedEOF)
	}
	copy(il.RepeatBits, b)
	b = b[len(il.RepeatBits):]
	// Iterate through bit stream and count bits (number of unique hashes)
	hash := &il.PrevHash
	var compressedCount int
	for _, repeatBits := range il.RepeatBits {
		compressedCount += 8 - bits.OnesCount8(repeatBits)
	}
	// Allocate space for unique hashes
	il.Compressed = make([][32]byte, compressedCount)
	// Iterate through bit stream and read hashes:
	// Bit 0: New Hash
	// Bit 1: Repeated Hash
	var compressedIndex int
	for i := uint(0); i < uint(count); i++ {
		repeated := il.RepeatBits[i/8]&(0x80>>(i%8)) != 0
		if !repeated {
			// Point to hash
			hash = &il.Compressed[compressedIndex]
			compressedIndex++
			// Read in data
			if len(b) < len(hash) {
				return 0, fmt.Errorf("failed to read hash: %w", beserial.ErrUnexpectedEOF)
			}
			b = b[len(hash):]
			copy(hash[:], b)
		}
		// Save pointer to last hash
		il.Hashes[i] = hash
	}
	return len(orig) - len(b), nil
}

func (il *BlockInterlink) MarshalBESerial(b []byte) ([]byte, error) {
	b = append(b, uint8(len(il.Hashes)))
	b = append(b, il.RepeatBits...)
	for i := 0; i < len(il.Compressed); i++ {
		b = append(b, il.Compressed[i][:]...)
	}
	return b, nil
}

func (il *BlockInterlink) SizeBESerial() (n int, err error) {
	n = 1 // hash count
	n += len(il.RepeatBits)
	n += len(il.Compressed) * 32
	return
}
