package wire

import (
	"fmt"

	"golang.org/x/crypto/blake2b"
	"terorie.dev/nimiq/beserial"
)

type SignatureProof struct {
	PublicKey  [32]byte
	MerklePath MerklePath
	Signature  [64]byte
}

func (s *SignatureProof) SignerAddress() (addr [20]byte) {
	merkleRoot := s.MerklePath.ComputeRoot(s.PublicKey[:])
	copy(addr[:], merkleRoot[:]) // Take leftmost bytes of merkle root
	return
}

type MerklePath struct {
	Branches BitSet // list of branches taken. zero = right, one = left.
	Hashes   [][32]byte
}

func (mp *MerklePath) UnmarshalBESerial(b []byte) (n int, err error) {
	orig := b
	// Read branch bits
	n, err = mp.Branches.UnmarshalBESerial(b)
	if err != nil {
		return
	}
	b = b[n:]
	// Read merkle path nodes
	mp.Hashes = make([][32]byte, mp.Branches.Len)
	for i := range mp.Hashes {
		// TODO Move length check out of loop
		if len(b) < 32 {
			return 0, fmt.Errorf("reading merkle path node %d: %w", i, beserial.ErrUnexpectedEOF)
		}
		copy(mp.Hashes[i][:], b)
		b = b[32:]
	}
	return len(orig) - len(b), nil
}

func (mp *MerklePath) MarshalBESerial(b []byte) ([]byte, error) {
	var err error
	b, err = mp.Branches.MarshalBESerial(b)
	if err != nil {
		return b, err
	}
	for _, hash := range mp.Hashes {
		b = append(b, hash[:]...)
	}
	return b, nil
}

func (mp *MerklePath) SizeBESerial() (n int, err error) {
	n, err = mp.Branches.SizeBESerial()
	if err != nil {
		return
	}
	n += len(mp.Hashes) * 32
	return
}

func (mp *MerklePath) ComputeRoot(leafValue []byte) (root [32]byte) {
	root = blake2b.Sum256(leafValue[:])
	for i, hash := range mp.Hashes {
		var node [64]byte
		if mp.Branches.bit(uint8(i)) {
			copy(node[:32], hash[:])
			copy(node[32:], root[:])
		} else {
			copy(node[:32], root[:])
			copy(node[32:], hash[:])
		}
		root = blake2b.Sum256(node[:])
	}
	return
}
