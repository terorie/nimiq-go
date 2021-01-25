// Package wallet contains utilities for handling a NIM wallet.
package wallet

import (
	"bytes"
	"crypto"
	"crypto/ed25519"

	"terorie.dev/nimiq/beserial"
	"terorie.dev/nimiq/wire"
)

// Basic is a minimal wallet, with a single keypair.
type Basic struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	address    [20]byte
}

// GenerateBasic generates a new Basic wallet with a random key.
func GenerateBasic() *Basic {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	b := &Basic{
		PublicKey:  pub,
		PrivateKey: priv,
	}
	return b
}

func (b *Basic) GetAddress() [20]byte {
	if b.address == [20]byte{} {
		pubKey := b.GetPublicKey()
		b.address = wire.PublicKeyToAddress(&pubKey)
	}
	return b.address
}

func (b *Basic) GetPublicKey() (pub [32]byte) {
	copy(pub[:], b.PublicKey)
	return
}

func (b *Basic) signTx(tx wire.Tx) []byte {
	content := tx.AsTxContent()
	contentBuf, err := beserial.Marshal(nil, &content)
	if err != nil {
		panic(err)
	}
	sig, _ := b.PrivateKey.Sign(nil, contentBuf, crypto.Hash(0))
	return sig
}

func (b *Basic) SignBasicTx(tx *wire.BasicTx) {
	sig := b.signTx(tx)
	copy(tx.Signature[:], sig)
}

func (b *Basic) SignExtendedTx(tx *wire.ExtendedTx) {
	sig := b.signTx(tx)
	var buf bytes.Buffer
	buf.Write(b.PublicKey[:32])
	// Empty merkle path
	_, _ = buf.Write([]byte{0x00})
	buf.Write(sig)
	tx.Proof = buf.Bytes()
}
