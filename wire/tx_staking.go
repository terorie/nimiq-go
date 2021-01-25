package wire

// StakingTxData is the extra data of a staking transaction.
// EXPERIMENTAL!
type StakingTxData struct {
	ValidatorKey     [96]byte
	RewardAddress    *[20]byte `beserial:"optional"`
	ProofOfKnowledge [48]byte
}
