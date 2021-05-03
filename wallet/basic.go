// Package wallet contains utilities for handling a NIM wallet.
package wallet

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"fmt"

	"terorie.dev/nimiq/beserial"
	"terorie.dev/nimiq/wire"
)

// Basic is a minimal wallet with a single keypair.
//
// It is implemented entirely in software with no security features whatsoever.
// Please don't use this with actual funds.
type Basic struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	address    [20]byte
}

// DeriveBasic creates a new Basic wallet from an Ed25519 private key.
func DeriveBasic(buf []byte) *Basic {
	priv := ed25519.NewKeyFromSeed(buf)
	pub := priv.Public().(ed25519.PublicKey)
	return &Basic{
		PrivateKey: priv,
		PublicKey:  pub,
	}
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

// GetAddress returns the address of the wallet.
func (b *Basic) GetAddress() [20]byte {
	if b.address == [20]byte{} {
		pubKey := b.GetPublicKey()
		b.address = wire.PublicKeyToAddress(&pubKey)
	}
	return b.address
}

// GetPublicKey returns the public key of the wallet.
func (b *Basic) GetPublicKey() (pub [32]byte) {
	if b.PublicKey == nil {
		panic("wallet not initialized")
	}
	copy(pub[:ed25519.PublicKeySize], b.PublicKey[:ed25519.PublicKeySize])
	return
}

func (b *Basic) signTx(tx wire.Tx) []byte {
	if b.PublicKey == nil {
		panic("wallet not initialized")
	}
	content := tx.AsTxContent()
	contentBuf, err := beserial.Marshal(nil, &content)
	if err != nil {
		panic(err)
	}
	sig, err := b.PrivateKey.Sign(nil, contentBuf, crypto.Hash(0))
	if err != nil {
		panic("ed25519.PrivateKey.Sign failed: " + err.Error())
	}
	return sig
}

// SignBasicTx signs a basic transaction.
func (b *Basic) SignBasicTx(tx *wire.BasicTx) error {
	if !bytes.Equal(b.PublicKey, tx.SenderPubKey[:]) {
		return fmt.Errorf("refusing to sign basic tx with a different sender")
	}
	sig := b.signTx(tx)
	copy(tx.Signature[:], sig)
	return nil
}

// SignExtendedTx signs an extended transaction.
func (b *Basic) SignExtendedTx(tx *wire.ExtendedTx) (err error) {
	sig := b.signTx(tx)
	var sigProof wire.SignatureProof
	copy(sigProof.PublicKey[:ed25519.PublicKeySize], b.PublicKey[:ed25519.PublicKeySize])
	copy(sigProof.Signature[:ed25519.SignatureSize], sig[:ed25519.SignatureSize])
	tx.Proof, err = beserial.Marshal(nil, &sigProof)
	return
}
