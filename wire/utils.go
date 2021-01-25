package wire

import (
	"math/big"

	"terorie.dev/nimiq/policy"
)

// TargetToCompact converts a target hash number into the compact "n-bits" representation.
func TargetToCompact(target big.Int) (compact uint32) {
	tgtBytes := target.Bytes()

	// If the first (most significant) byte is
	// greater than 127 (0x7f), prepend a zero byte.
	if tgtBytes[0] >= 0x80 && len(tgtBytes) >= 3 {
		compact |= uint32(len(tgtBytes)+1) << 24
		compact |= uint32(tgtBytes[0]) << 8
		compact |= uint32(tgtBytes[1])
	} else {
		compact |= uint32(len(tgtBytes)) << 24
		compact |= uint32(tgtBytes[0]) << 16
		compact |= uint32(tgtBytes[1]) << 8
		compact |= uint32(tgtBytes[2])
	}

	return
}

// DifficultyToCompact converts a difficulty number into the compact "n-bits" representation.
func DifficultyToCompact(difficulty big.Int) uint32 {
	return TargetToCompact(DifficultyToTarget(difficulty))
}

// DifficultyToTarget converts a difficulty number into the target hash number.
func DifficultyToTarget(difficulty big.Int) (target big.Int) {
	target.Div(&policy.BlockTargetMax, &difficulty)
	return
}
