package tree

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terorie.dev/nimiq/wire"
)

func TestTree_Empty(t *testing.T) {
	tree := PMTree{Store: NewMemStore()}
	hash := tree.Hash()
	assert.Equal(t, "ab29e6dc16755d0071eba349ebda225d15e4f910cb474549c47e95cb85ecc4d6", hex.EncodeToString(hash[:]))
}

func TestTree_PutAndGet(t *testing.T) {
	tree := PMTree{Store: NewMemStore()}

	address1 := &[20]byte{}
	account1 := wire.WrapAccount{
		Account: &wire.BasicAccount{Value: 5},
	}
	account1Blob, _ := account1.MarshalBESerial(nil)
	tree.PutEntry(address1, account1Blob)
	rootHash1 := tree.Hash()
	assert.Equal(t,
		"4644a8c8bc0b333230751e6fbcd0b49f4ebc7bb682df3a97e091115d0fe26c05",
		hex.EncodeToString(rootHash1[:]))

	address2 := &[20]byte{0x10}
	account2 := wire.WrapAccount{
		Account: &wire.BasicAccount{Value: 55},
	}
	account2Blob, _ := account2.MarshalBESerial(nil)
	tree.PutEntry(address2, account2Blob)
	rootHash2 := tree.Hash()
	assert.Equal(t,
		"f6fc7ecf89d94fa4e91a19805d36976ccd09633a95ff201981167e8c68c141dd",
		hex.EncodeToString(rootHash2[:]))

	address3 := &[20]byte{0x12}
	account3 := wire.WrapAccount{
		Account: &wire.BasicAccount{Value: 55555555},
	}
	account3Blob, _ := account3.MarshalBESerial(nil)
	tree.PutEntry(address3, account3Blob)
	rootHash3 := tree.Hash()
	assert.Equal(t,
		"c8a459ea666e3b027dbef89c00e7600d22c0a0c7ff8051a5e9687026d027c0f5",
		hex.EncodeToString(rootHash3[:]))
}

func BenchmarkPMTree_PutEntry(b *testing.B) {
	tree := &PMTree{Store: NewMemStore()}
	for i := 0; i < b.N; i++ {
		var addr [20]byte
		_, _ = rand.Read(addr[:])
		tree.PutEntry(&addr, []byte("0123"))
	}
}

func TestPMTree_InsertionOrderInvariance(t *testing.T) {
	tree := &PMTree{Store: NewMemStore()}

	// A few accounts
	address1 := &[20]byte{}
	account1 := wire.WrapAccount{Account: &wire.BasicAccount{Value: 5}}
	account1Buf, err := account1.MarshalBESerial(nil)
	require.NoError(t, err)
	address2 := &[20]byte{0x10}
	account2 := wire.WrapAccount{Account: &wire.BasicAccount{Value: 55}}
	account2Buf, err := account2.MarshalBESerial(nil)
	require.NoError(t, err)
	address3 := &[20]byte{0x12}
	account3 := wire.WrapAccount{Account: &wire.BasicAccount{Value: 55555555}}
	account3Buf, err := account3.MarshalBESerial(nil)
	require.NoError(t, err)

	const emptyHash = "ab29e6dc16755d0071eba349ebda225d15e4f910cb474549c47e95cb85ecc4d6"
	const actualHash = "c8a459ea666e3b027dbef89c00e7600d22c0a0c7ff8051a5e9687026d027c0f5"

	root0Empty := tree.Hash()
	assert.Equal(t, emptyHash, hex.EncodeToString(root0Empty[:]))

	// Order 1
	assert.True(t, tree.PutEntry(address1, account1Buf))
	assert.True(t, tree.PutEntry(address2, account2Buf))
	assert.True(t, tree.PutEntry(address3, account3Buf))
	root1 := tree.Hash()
	assert.Equal(t, actualHash, hex.EncodeToString(root1[:]))

	// Reset
	assert.True(t, tree.PutEntry(address1, nil))
	assert.True(t, tree.PutEntry(address2, nil))
	assert.True(t, tree.PutEntry(address3, nil))
	empty1Hash := tree.Hash()
	assert.Equal(t, emptyHash, hex.EncodeToString(empty1Hash[:]))

	// Order 2
	assert.True(t, tree.PutEntry(address1, account1Buf))
	assert.True(t, tree.PutEntry(address3, account3Buf))
	assert.True(t, tree.PutEntry(address2, account2Buf))
	root2 := tree.Hash()
	assert.Equal(t, actualHash, hex.EncodeToString(root2[:]))

	// Reset
	assert.True(t, tree.PutEntry(address1, nil))
	assert.True(t, tree.PutEntry(address2, nil))
	assert.True(t, tree.PutEntry(address3, nil))
	empty2Hash := tree.Hash()
	assert.Equal(t, emptyHash, hex.EncodeToString(empty2Hash[:]))

	// Order 3
	assert.True(t, tree.PutEntry(address2, account2Buf))
	assert.True(t, tree.PutEntry(address1, account1Buf))
	assert.True(t, tree.PutEntry(address3, account3Buf))
	root3 := tree.Hash()
	assert.Equal(t, actualHash, hex.EncodeToString(root3[:]))

	// Reset
	assert.True(t, tree.PutEntry(address1, nil))
	assert.True(t, tree.PutEntry(address2, nil))
	assert.True(t, tree.PutEntry(address3, nil))
	empty3Hash := tree.Hash()
	assert.Equal(t, emptyHash, hex.EncodeToString(empty3Hash[:]))
}
