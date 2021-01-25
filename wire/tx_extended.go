package wire

// ExtendedTx is like a BasicTx but allows non-basic sender & recipients,
// also non-nil data and complicated proofs.
type ExtendedTx struct {
	Sender              [20]byte
	SenderType          uint8
	Recipient           [20]byte
	RecipientType       uint8
	Value               uint64
	Fee                 uint64
	ValidityStartHeight uint32
	Flags               uint8
	Data                []byte
	Proof               []byte
	NetworkID           uint8
}

// Type returns TxExtended.
func (t *ExtendedTx) Type() uint8 {
	return TxExtended
}

func (t *ExtendedTx) Amount() (value uint64, fee uint64) {
	return t.Value, t.Fee
}

// AsExtendedTx returns itself - it's already an ExtendedTx.
func (t *ExtendedTx) AsExtendedTx() ExtendedTx {
	return *t
}

// AsTxContent returns the simplified transaction content of the extended tx.
func (t *ExtendedTx) AsTxContent() TxContent {
	return TxContent{
		Data:                t.Data,
		Sender:              t.Sender,
		SenderType:          t.SenderType,
		Recipient:           t.Recipient,
		RecipientType:       t.RecipientType,
		Value:               t.Value,
		Fee:                 t.Fee,
		ValidityStartHeight: t.ValidityStartHeight,
		NetworkID:           t.NetworkID,
		Flags:               t.Flags,
	}
}

func (t *ExtendedTx) GetSender() (addr *[20]byte, typ uint8) {
	return &t.Sender, t.SenderType
}

func (t *ExtendedTx) GetRecipient() (addr *[20]byte, typ uint8) {
	return &t.Recipient, t.RecipientType
}

var _ Tx = (*ExtendedTx)(nil)
