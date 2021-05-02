package tree

import (
	"io"

	"golang.org/x/crypto/blake2b"
)

// Node is an entry in the state tree.
type Node interface {
	GetPrefix() Nibbles
}

// Branch is a branch node in the tree with 16 children.
type Branch struct {
	Prefix   Nibbles
	Children [16]Child
}

// PutChild sets the child to the hash of another node.
func (n *Branch) PutChild(suffix Nibbles, hash [32]byte) {
	n.Children[suffix[0]] = Child{
		Suffix: suffix,
		Hash:   hash,
		Exists: true,
	}
}

// GetPrefix returns the partial path of the node.
func (n *Branch) GetPrefix() Nibbles {
	return n.Prefix
}

// Hash calculates the tree node hash.
func (n *Branch) Hash() (sum [32]byte) {
	h, _ := blake2b.New256(nil)
	h.Write([]byte{0x00, uint8(len(n.Prefix))})
	h.Write(n.Prefix.HexBytes())
	var childCount uint8
	for _, child := range n.Children {
		if child.Exists {
			childCount++
		}
	}
	h.Write([]byte{childCount})
	for _, child := range n.Children {
		if !child.Exists {
			continue
		}
		h.Write([]byte{uint8(len(child.Suffix))})
		h.Write(child.Suffix.HexBytes())
		h.Write(child.Hash[:])
	}
	h.Sum(sum[:0])
	return
}

// Child is a reference to another node, embedded in a branch node.
type Child struct {
	Suffix Nibbles
	Hash   [32]byte
	Exists bool
}

// Leaf is a leaf node containing an account.
type Leaf struct {
	Prefix Nibbles
	Value  []byte
}

// GetPrefix returns the partial path of the node.
func (t *Leaf) GetPrefix() Nibbles {
	return t.Prefix
}

func (t *Leaf) MarshalTo(w io.Writer) {
	_, _ = w.Write([]byte{0xFF, 40})
	_, _ = w.Write(t.Prefix.HexBytes())
	_, _ = w.Write(t.Value)
}

// Hash calculates the leaf node hash.
func (t *Leaf) Hash() (sum [32]byte) {
	h, _ := blake2b.New256(nil)
	t.MarshalTo(h)
	h.Sum(sum[:0])
	return
}
