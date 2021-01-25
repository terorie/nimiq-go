package wire

import (
	"encoding/binary"
	"fmt"

	"terorie.dev/nimiq/beserial"
)

// Message IDs.
const (
	MessageVersion   = 0
	MessageInv       = 1
	MessageGetData   = 2
	MessageGetHeader = 3
	MessageNotFound  = 4
	MessageGetBlocks = 5
	MessageBlock     = 6
	MessageHeader    = 7
	MessageTx        = 8
	MessageMempool   = 9
	MessageReject    = 10
	MessageSubscribe = 11

	MessageAddr    = 20
	MessageGetAddr = 21
	MessagePing    = 22
	MessagePong    = 23

	MessageSignal = 30

	MessageGetChainProof        = 40
	MessageChainProof           = 41
	MessageGetAccountsProof     = 42
	MessageAccountsProof        = 43
	MessageGetAccountsTreeChunk = 44
	MessageAccountsTreeChunk    = 45
	MessageGetTxProof           = 47
	MessageTxProof              = 48
	MessageGetTxReceipts        = 49
	MessageTxReceipts           = 50
	MessageGetBlockProof        = 51
	MessageBlockProof           = 52

	MessageGetHead = 60
	MessageHead    = 61

	MessageVerAck = 90
)

// UnmarshalMessage creates the message with the specified type
// from the beserial-encoded buffer.
// Unlike how other beserial methods behave, the provided buffer
// must not have data past the end of the message, or an error is returned.
func UnmarshalMessage(msgType uint64, buf []byte) (m Message, err error) {
	switch msgType {
	case MessageVersion:
		m = new(VersionMessage)
	case MessageInv:
		m = &InvMessage{MessageType: MessageInv}
	case MessageGetData:
		m = &InvMessage{MessageType: MessageGetData}
	case MessageGetHeader:
		m = &InvMessage{MessageType: MessageGetHeader}
	case MessageNotFound:
		m = &InvMessage{MessageType: MessageNotFound}
	case MessageGetBlocks:
		m = new(GetBlocksMessage)
	case MessageBlock:
		m = new(BlockMessage)
	case MessageHeader:
		m = new(HeaderMessage)
	case MessageTx:
		m = new(TxMessage)
	case MessageMempool:
		m = MempoolMessage
	case MessageAddr:
		m = new(AddrMessage)
	case MessageGetAddr:
		m = new(GetAddrMessage)
	case MessagePing:
		m = &BasePingMessage{Pong: false}
	case MessagePong:
		m = &BasePingMessage{Pong: true}
	case MessageGetChainProof:
		m = GetChainProofMessage
	case MessageGetHead:
		m = GetHeadMessage
	case MessageHead:
		m = new(HeadMessage)
	case MessageVerAck:
		m = new(VerAckMessage)
	default:
		return nil, UnknownMessageError(msgType)
	}
	var n int
	n, err = beserial.Unmarshal(buf, m)
	if err != nil {
		return nil, err
	}
	if n != len(buf) {
		return nil, fmt.Errorf("got %d bytes of junk after message", len(buf)-n)
	}
	return
}

// MarshalMessage encodes the message to beserial and prefixes the message type ID.
func MarshalMessage(m Message) (final []byte, err error) {
	size, err := beserial.Size(m)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 8+size)
	binary.BigEndian.PutUint64(buf[:8], m.Type())
	final, err = beserial.Marshal(buf[:8], m)
	return
}

// EmptyMessage is used when the message ID is enough.
// Take the "GetHead" message for example:
// It's entire purpose is in the name: Request the current blockchain head.
type EmptyMessage uint64

func (_ EmptyMessage) UnmarshalBESerial(_ []byte) (int, error) {
	return 0, nil
}

func (_ EmptyMessage) MarshalBESerial(b []byte) ([]byte, error) {
	return b, nil
}

func (_ EmptyMessage) SizeBESerial() (int, error) {
	return 0, nil
}

func (m EmptyMessage) Type() uint64 {
	return uint64(m)
}

// Empty messages
var (
	MempoolMessage       = EmptyMessage(MessageMempool)
	GetChainProofMessage = EmptyMessage(MessageGetChainProof)
	GetHeadMessage       = EmptyMessage(MessageGetHead)
)

// UnknownMessageError is returned when no decoder is available
// for the received message. Despite the name, it's not exactly
// an error, and the node should just ignore the message.
type UnknownMessageError int64

func (e UnknownMessageError) Error() string {
	return fmt.Sprintf("unknown message type: 0x%x", int64(e))
}
