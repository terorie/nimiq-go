package wire

type AccountsTreeChunk struct {
	Nodes [][]byte `beserial:"len_tag=uint16"`
	Proof AccountsProof
}

type AccountsProof struct {
	Nodes [][]byte `beserial:"len_tag=uint16"`
}
