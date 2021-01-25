package wire

type InvVector struct {
	Type uint32
	Hash [32]byte
}

const (
	InvError = 0
	InvTx    = 1
	InvBlock = 2
)

const VectorsMaxCount = 1000

type InvMessage struct {
	MessageType uint64
	Vectors     []InvVector
}

func (m *InvMessage) Type() uint64 {
	return m.MessageType
}

type GetAddrMessage struct {
	ProtocolMask uint8
	ServiceMask  uint32
	MaxResults   uint16
}

func (m *GetAddrMessage) Type() uint64 {
	return MessageGetAddr
}

type GetBlocksMessage struct {
	Locators   [][32]byte `beserial:"len_tag=uint16"`
	MaxInvSize uint16
	Direction  uint8
}

const (
	DirectionForward  = 1
	DirectionBackward = 2
)

func (m *GetBlocksMessage) Type() uint64 {
	return MessageGetBlocks
}

type GetTxProofMessage struct {
	BlockHash [32]byte
	Addresses [][20]byte
}

func (m *GetTxProofMessage) Type() uint64 {
	return MessageGetTxProof
}

type GetTxReceiptsMessage struct {
	Address [20]byte
	Offset  uint32
}

func (m *GetTxReceiptsMessage) Type() uint64 {
	return MessageGetTxReceipts
}

type BlockMessage struct {
	Block
}

func (m *BlockMessage) Type() uint64 {
	return MessageBlock
}

type HeadMessage struct {
	BlockHeader
}

func (m *HeadMessage) Type() uint64 {
	return MessageHead
}

type HeaderMessage struct {
	BlockHeader
}

func (m *HeaderMessage) Type() uint64 {
	return MessageBlock
}
