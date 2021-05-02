package wire

import (
	"fmt"

	"terorie.dev/nimiq/beserial"
)

// Tx is implemented by all on-chain transactions (basic & extended format).
type Tx interface {
	Amount() (value uint64, fee uint64)
	Type() uint8
	AsTxContent() TxContent
	AsExtendedTx() ExtendedTx
	GetSender() (addr *[20]byte, typ uint8)
	GetRecipient() (addr *[20]byte, typ uint8)
}

// WrapTx implements BESerial for Tx.
type WrapTx struct {
	Tx Tx
}

// Transaction types.
const (
	TxBasic    = uint8(0)
	TxExtended = uint8(1)
)

// Transaction flags.
const (
	TxFlagNone             = uint8(0x00)
	TxFlagContractCreation = uint8(0x01)
)

// UnmarshalBESerial decodes the transaction from beserial,
// using a type prefix to choose the transaction type.
func (wt *WrapTx) UnmarshalBESerial(buf []byte) (n int, err error) {
	if len(buf) < 1 {
		return 0, beserial.ErrUnexpectedEOF
	}
	txType := buf[0]
	n++
	buf = buf[1:]
	switch txType {
	case TxBasic:
		wt.Tx = new(BasicTx)
	case TxExtended:
		wt.Tx = new(ExtendedTx)
	default:
		return 0, fmt.Errorf("invalid tx type: %d", txType)
	}
	var sub int
	sub, err = beserial.Unmarshal(buf, wt.Tx)
	n += sub
	return
}

// MarshalBESerial encodes the transaction to beserial with a type prefix.
func (wt WrapTx) MarshalBESerial(b []byte) ([]byte, error) {
	b = append(b, wt.Tx.Type())
	return beserial.Marshal(b, wt.Tx)
}

// SizeBESerial returns the size of the type-prefixed transaction.
func (wt WrapTx) SizeBESerial() (n int, err error) {
	n = 1
	var sub int
	sub, err = beserial.Size(wt.Tx)
	n += sub
	return
}

// TxContent is a simplified representation of a transaction.
// It contains all common fields but omits some details.
// Also, it's used for signing, implying for omitted fields are not security-relevant.
type TxContent struct {
	Data                []byte `beserial:"len_tag=uint16"`
	Sender              [20]byte
	SenderType          uint8
	Recipient           [20]byte
	RecipientType       uint8
	Value               uint64
	Fee                 uint64
	ValidityStartHeight uint32
	NetworkID           uint8
	Flags               uint8
}
