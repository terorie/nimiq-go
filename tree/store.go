package tree

// Store is the low-level key-value storage backend
type Store interface {
	GetNode(Nibbles) Node
	PutNode(Nibbles, Node)
	DelNode(Nibbles)
}

// Nibbles is a list of 4-bit segments.
// The list is uncompressed, i.e. each byte has only the lower 4 bit set.
type Nibbles []byte

const hexTable = "0123456789abcdef"

// KeyToNibbles expands a 20-byte address to a list of 40 nibbles.
func KeyToNibbles(key *[20]byte) Nibbles {
	nb := make([]byte, 40)
	for i, b := range key {
		nb[2*i] = b >> 4
		nb[(2*i)+1] = b & 0xF
	}
	return nb
}

// ToKey compacts a list of 40 nibbles to a 20-byte address.
// Panics if used with not exactly 40 nibbles.
func (nbs Nibbles) ToKey() (key [20]byte) {
	if len(nbs) != 40 {
		panic("ToKey called on len != 40")
	}
	for i := 0; i < 40; i += 2 {
		key[i/2] = (nbs[i] << 4) | nbs[i+1]
	}
	return
}

func (nbs Nibbles) Compress() []byte {
	buf := make([]byte, (len(nbs)-1)/2)
	for i, n := range nbs {
		if i%2 == 0 {
			buf[i/2] = n << 4
		} else {
			buf[i/2] |= n & 0xF
		}
	}
	return buf
}

// HexBytes serializes the nibbles to a hex string.
func (nbs Nibbles) HexBytes() []byte {
	hexBuf := make([]byte, len(nbs))
	for i, nb := range nbs {
		hexBuf[i] = hexTable[nb]
	}
	return hexBuf
}

func (nbs Nibbles) String() string {
	return string(nbs.HexBytes())
}

func (nbs Nibbles) CommonPrefix(onbs Nibbles) (common Nibbles) {
	for i := range nbs {
		if i >= len(onbs) {
			return
		}
		if nbs[i] == onbs[i] {
			common = nbs[:i+1]
		} else {
			return
		}
	}
	return
}

func (nbs Nibbles) PrefixOf(onbs Nibbles) bool {
	if len(nbs) > len(onbs) {
		return false
	}
	for i := range nbs {
		if nbs[i] != onbs[i] {
			return false
		}
	}
	return true
}
