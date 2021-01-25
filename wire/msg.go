package wire

type Message interface {
	Type() uint64
}

type AddrMessage struct {
	Addresses []PeerAddress `beserial:"len_tag=uint16"`
}

func (m *AddrMessage) Type() uint64 {
	return MessageAddr
}

type TxMessage struct {
	Tx WrapTx
	//Proof *AccountsProof `beserial:"optional"`
}

func (m *TxMessage) Type() uint64 {
	return MessageTx
}
