package wire

import (
	"encoding/binary"

	"terorie.dev/nimiq/beserial"
)

// Generic peer-to-peer messages and handshake stuff.

// VersionMessage is the first message of a Nimiq handshake.
type VersionMessage struct {
	Version        uint32
	PeerAddress    PeerAddress
	GenesisHash    [32]byte
	HeadHash       [32]byte
	ChallengeNonce [32]byte
	UserAgent      string `beserial:"len_tag=uint8"`
}

func (m *VersionMessage) Type() uint64 {
	return MessageVersion
}

// VerAckMessage is the second message of a Nimiq handshake.
type VerAckMessage struct {
	PublicKey [32]byte
	Signature [64]byte
}

func (m *VerAckMessage) Type() uint64 {
	return MessageVerAck
}

// BasePingMessage is a ping or pong message.
// It implements the heartbeat mechanism.
type BasePingMessage struct {
	Nonce uint32
	Pong  bool
}

func (p *BasePingMessage) Type() uint64 {
	if !p.Pong {
		return MessagePing
	} else {
		return MessagePong
	}
}

// UnmarshalBESerial unmarshals the ping message content.
func (p *BasePingMessage) UnmarshalBESerial(buf []byte) (n int, err error) {
	if len(buf) < 4 {
		return 0, beserial.ErrUnexpectedEOF
	}
	p.Nonce = binary.BigEndian.Uint32(buf)
	return 4, nil
}

// MarshalBESerial marshals the ping message content.
func (p *BasePingMessage) MarshalBESerial(b []byte) ([]byte, error) {
	var num [4]byte
	binary.BigEndian.PutUint32(num[:], p.Nonce)
	b = append(b, num[:]...)
	return b, nil
}

// SizeBESerial returns the size of the ping message content, which is always 4.
func (p *BasePingMessage) SizeBESerial() (n int, err error) {
	return 4, nil
}
