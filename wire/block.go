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
	Repeats    BitSet
	Hashes     []*[32]byte // pointers to Compressed
	PrevHash   [32]byte
	Compressed [][32]byte
}

// UnmarshalBESerial implements the interlink decoding algorithm.
func (il *BlockInterlink) UnmarshalBESerial(b []byte) (n int, err error) {
	orig := b
	// Read repeat bits
	n, err = il.Repeats.UnmarshalBESerial(b)
	if err != nil {
		return
	}
	b = b[n:]
	// Iterate through bit stream and count bits (number of unique hashes)
	il.Hashes = make([]*[32]byte, il.Repeats.Len)
	hash := &il.PrevHash
	var compressedCount int
	for _, repeatBits := range il.Repeats.Bits {
		compressedCount += 8 - bits.OnesCount8(repeatBits)
	}
	// Allocate space for unique hashes
	il.Compressed = make([][32]byte, compressedCount)
	// Iterate through bit stream and read hashes:
	// Bit 0: New Hash
	// Bit 1: Repeated Hash
	var compressedIndex int
	for i := uint8(0); i < il.Repeats.Len; i++ {
		if !il.Repeats.bit(i) {
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
	var err error
	b, err = il.Repeats.MarshalBESerial(b)
	if err != nil {
		return b, err
	}
	for i := 0; i < len(il.Compressed); i++ {
		b = append(b, il.Compressed[i][:]...)
	}
	return b, nil
}

func (il *BlockInterlink) SizeBESerial() (n int, err error) {
	n, err = il.Repeats.SizeBESerial()
	if err != nil {
		return
	}
	n += len(il.Compressed) * 32
	return
}
