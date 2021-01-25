package wire

// BasicTx is a simple transaction moving NIM from one account to another.
type BasicTx struct {
	SenderPubKey        [32]byte
	Recipient           [20]byte
	Value               uint64
	Fee                 uint64
	ValidityStartHeight uint32
	NetworkID           uint8
	Signature           [64]byte
}

// Type returns TxBasic.
func (t *BasicTx) Type() uint8 {
	return TxBasic
}

func (t *BasicTx) Amount() (value uint64, fee uint64) {
	return t.Value, t.Fee
}

// AsExtendedTx returns the extended equivalent of this basic tx.
func (t *BasicTx) AsExtendedTx() ExtendedTx {
	return ExtendedTx{
		Sender:              PublicKeyToAddress(&t.SenderPubKey),
		SenderType:          AccountBasic,
		Recipient:           t.Recipient,
		RecipientType:       AccountBasic,
		Value:               t.Value,
		Fee:                 t.Fee,
		ValidityStartHeight: t.ValidityStartHeight,
		Flags:               0,
		Data:                nil,
		Proof:               t.Signature[:],
		NetworkID:           t.NetworkID,
	}
}

// AsTxContent returns the simplified transaction content.
func (t *BasicTx) AsTxContent() TxContent {
	return TxContent{
		Data:                nil,
		Sender:              PublicKeyToAddress(&t.SenderPubKey),
		SenderType:          AccountBasic,
		Recipient:           t.Recipient,
		RecipientType:       AccountBasic,
		Value:               t.Value,
		Fee:                 t.Fee,
		ValidityStartHeight: t.ValidityStartHeight,
		NetworkID:           t.NetworkID,
		Flags:               0,
	}
}

func (t *BasicTx) GetSender() (addr *[20]byte, typ uint8) {
	addr = new([20]byte)
	*addr = PublicKeyToAddress(&t.SenderPubKey)
	typ = AccountBasic
	return
}

func (t *BasicTx) GetRecipient() (addr *[20]byte, typ uint8) {
	return &t.Recipient, AccountBasic
}

var _ Tx = (*BasicTx)(nil)
