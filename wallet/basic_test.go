package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terorie.dev/nimiq/wire"
)

func TestBasic_SignBasicTx(t *testing.T) {
	var w Basic
	w.PrivateKey = []byte{
		0x48, 0x34, 0x2c, 0xab, 0x09, 0xb5, 0x6e, 0xfe, 0xf8, 0x24, 0x34, 0x4b, 0x89, 0x32, 0x3c, 0xb7,
		0x75, 0xb3, 0xb7, 0xc2, 0x1f, 0xec, 0x91, 0x16, 0xf0, 0xdf, 0xda, 0x83, 0xd1, 0x93, 0xb5, 0x0b,
	}
	tx := wire.BasicTx{
		SenderPubKey:        w.GetPublicKey(),
		Recipient:           [20]byte{},
		Value:               420,
		Fee:                 1337,
		ValidityStartHeight: 99,
		NetworkID:           42,
	}
	w.SignBasicTx(&tx)

	txBuf, err := wire.WrapTx{Tx: &tx}.MarshalBESerial(nil)
	require.NoError(t, err)
	assert.Equal(t,
		"002a39f666099582d659112b8c196630958095805f6b1016400030ed0b182064d1000000000000000000000000000000000000000000000000000001a40000000000000539000000632a6a6d2fad4a136a9409aeb935c62af7a356e86435829d94f8a97ddc5fc5e5e202ffedfdf7dc438cc48dd5883edc9b07a2971cafaca520cb34f00a3eb4834be60f",
		hex.EncodeToString(txBuf))
}

func TestBasic_SignExtendedTx(t *testing.T) {
	var w Basic
	w.PrivateKey = []byte{
		0x48, 0x34, 0x2c, 0xab, 0x09, 0xb5, 0x6e, 0xfe, 0xf8, 0x24, 0x34, 0x4b, 0x89, 0x32, 0x3c, 0xb7,
		0x75, 0xb3, 0xb7, 0xc2, 0x1f, 0xec, 0x91, 0x16, 0xf0, 0xdf, 0xda, 0x83, 0xd1, 0x93, 0xb5, 0x0b,
	}
	tx := wire.ExtendedTx{
		Sender:              w.address,
		SenderType:          0,
		Recipient:           [20]byte{},
		RecipientType:       0,
		Value:               1337,
		Fee:                 420,
		ValidityStartHeight: 99,
		Flags:               0,
		Data:                nil,
		NetworkID:           42,
	}
	w.SignExtendedTx(&tx)

	txBuf, err := wire.WrapTx{Tx: &tx}.MarshalBESerial(nil)
	require.NoError(t, err)
	assert.Equal(t,
		"010000a1f5245a31b48f1f0e4839c2b1b260a89317e85500000000000000000000000000000000000000000000000000000000053900000000000001a4000000632a0000612a39f666099582d659112b8c196630958095805f6b1016400030ed0b182064d1002cf0150ce3fddd6d3d70de801a5e0744e8b3fc946a4380cc134127015e32b13580accc717c7af05e0334fe7ada1e99920e7dd1e03a45091db3986c08c977fc09",
		hex.EncodeToString(txBuf))
}
