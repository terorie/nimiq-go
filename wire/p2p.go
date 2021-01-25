package wire

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"net"

	"golang.org/x/crypto/blake2b"
	"terorie.dev/nimiq/beserial"
)

// NetAddress is an IPv4, IPv6, or placeholder address.
type NetAddress struct {
	Type uint8
	net.IP
}

// Types of NetAddress.
const (
	NetAddressIPv4        = 0
	NetAddressIPv6        = 1
	NetAddressUnspecified = 2
	NetAddressUnknown     = 3
)

func (na *NetAddress) UnmarshalBESerial(b []byte) (n int, err error) {
	var t uint8
	n, err = beserial.Unmarshal(b, &t)
	if err != nil {
		return
	}
	b = b[n:]
	switch t {
	case NetAddressIPv4:
		na.IP = make([]byte, net.IPv4len)
	case NetAddressIPv6:
		na.IP = make([]byte, net.IPv6len)
	default:
		return n, fmt.Errorf("invalid NetAddress type: 0x%x", t)
	}
	if len(b) < len(na.IP) {
		return n, beserial.ErrUnexpectedEOF
	}
	n += copy(na.IP, b)
	return
}

func (na *NetAddress) MarshalBESerial(b []byte) ([]byte, error) {
	b = append(b, na.Type)
	b = append(b, na.IP...)
	return b, nil
}

func (na *NetAddress) SizeBESerial() (int, error) {
	return na.sizeBESerial(), nil
}

func (na *NetAddress) sizeBESerial() int {
	return 1 + len(na.IP)
}

// PeerAddress contains an overview about a peer and how to reach it.
type PeerAddress struct {
	PeerAddressHeader
	Detail interface{}
}

// PeerAddressHeader contains the protocol-agnostic peer address details.
type PeerAddressHeader struct {
	Protocol   uint8
	Services   uint32
	Timestamp  uint64
	NetAddress NetAddress
	PublicKey  [32]byte
	Distance   uint8
	Signature  [64]byte
}

// Types of transport protocols.
const (
	ProtocolDumb = uint8(0)
	ProtocolWSS  = uint8(1 << 0)
	ProtocolRTC  = uint8(1 << 1)
	ProtocolWS   = uint8(1 << 2)
)

// ProtocolScheme returns the URI scheme of a protocol ID.
func ProtocolScheme(proto uint8) string {
	switch proto {
	case ProtocolDumb:
		return "dumb"
	case ProtocolWS:
		return "ws"
	case ProtocolWSS:
		return "wss"
	case ProtocolRTC:
		return "rtc"
	default:
		return "invalid"
	}
}

func (pa *PeerAddress) UnmarshalBESerial(b []byte) (n int, err error) {
	var sub int
	sub, err = beserial.Unmarshal(b, &pa.PeerAddressHeader)
	n += sub
	b = b[sub:]
	if err != nil {
		return
	}
	switch pa.Protocol {
	case ProtocolDumb:
		pa.Detail = nil
	case ProtocolWS, ProtocolWSS:
		pa.Detail = new(PeerAddressSrv)
	default:
		return n, fmt.Errorf("unknown protocol: 0x%x", pa.Protocol)
	}
	sub, err = beserial.Unmarshal(b, &pa.Detail)
	n += sub
	b = b[sub:]
	if err != nil {
		return
	}
	return
}

func (pa *PeerAddress) MarshalBESerial(b []byte) ([]byte, error) {
	var err error
	b, err = beserial.Marshal(b, &pa.PeerAddressHeader)
	if err != nil {
		return nil, err
	}
	return beserial.Marshal(b, &pa.Detail)
}

func (pa *PeerAddress) SizeBESerial() (n int, err error) {
	var sub int
	sub, err = beserial.Size(&pa.PeerAddressHeader)
	if err != nil {
		return 0, err
	}
	n += sub
	sub, err = beserial.Size(&pa.Detail)
	if err != nil {
		return 0, err
	}
	n += sub
	return n, nil
}

// PeerID is an 128-bit identifier of a peer.
type PeerID [16]byte

// GetPeerID derives a peer ID from the node public key.
func GetPeerID(pubKey ed25519.PublicKey) (p PeerID) {
	hash := blake2b.Sum256(pubKey)
	copy(p[:], hash[:16])
	return
}

// Sign signs the protocol, services and timestamp fields
// and puts the resulting Ed25519 in the struct.
func (pa *PeerAddress) Sign(pk ed25519.PrivateKey) {
	var buf [13]byte
	buf[0] = pa.Protocol
	binary.BigEndian.PutUint32(buf[1:], pa.Services)
	binary.BigEndian.PutUint64(buf[5:], pa.Timestamp)
	signature := ed25519.Sign(pk, buf[:])
	copy(pa.PublicKey[:], pk[32:])
	copy(pa.Signature[:], signature[:])
}

// VerifySignature checks whether the protocol, services
// and timestamp fields have been signed correctly.
func (pa *PeerAddress) VerifySignature() bool {
	var buf [13]byte
	buf[0] = pa.Protocol
	binary.BigEndian.PutUint32(buf[1:], pa.Services)
	binary.BigEndian.PutUint32(buf[1:], pa.Services)
	binary.BigEndian.PutUint64(buf[5:], pa.Timestamp)
	return ed25519.Verify(pa.PublicKey[:], buf[:], pa.Signature[:])
}

// PeerAddressSrv is a client-server type address used
// for describing WS/WSS and TCP/TLS service locations.
type PeerAddressSrv struct {
	Host string `beserial:"len_tag=uint8"`
	Port uint16
	TLS  bool
}
