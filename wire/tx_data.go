package wire

type VestingCreation struct {
	Beneficiary [20]byte
}

type vestingMinimalCreation struct {
	Balance           uint64
	VestingStepBlocks uint32
}

type vestingNormalCreation struct {
	Balance           uint64
	VestingStart      uint32
	VestingStepBlocks uint32
	VestingStepAmount uint64
}

type vestingFullCreation struct {
	Balance            uint64
	VestingStart       uint32
	VestingStepBlocks  uint32
	VestingStepAmount  uint64
	VestingTotalAmount uint64
}
