// Package policy handles constant chain parameters.
package policy

import (
	"fmt"
	"math/big"
	"sync"
)

var BlockTargetMax big.Int

func init() {
	// 1 << 240
	BlockTargetMax.SetBytes([]byte{
		0x01,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
	})
}

func BlockReward(height uint32) uint64 {
	if height < 1 {
		panic(fmt.Sprintf("invalid height: %d", height))
	}
	supply := SupplyAfter(height - 1)
	return BlockRewardBySupply(supply, height)
}

const SupplyCacheInterval = uint32(5000)

var supplyCache = []uint64{InitialSupply}
var supplyCacheLock sync.Mutex

func SupplyAfter(height uint32) uint64 {
	endI := height / SupplyCacheInterval

	supplyCacheLock.Lock()
	var startI uint32
	if len(supplyCache) == 0 {
		startI = 0
	} else if uint32(len(supplyCache)-1) < endI {
		startI = uint32(len(supplyCache) - 1)
	} else {
		startI = endI
	}
	supply := supplyCache[startI]
	for i := startI; i < endI; i++ {
		startHeight := i * SupplyCacheInterval
		endHeight := startHeight + SupplyCacheInterval
		supply = SupplyBetween(supply, startHeight, endHeight)
		supplyCache = append(supplyCache, supply)
	}
	supplyCacheLock.Unlock()

	return SupplyBetween(supply, endI*SupplyCacheInterval, height+1)
}

func SupplyBetween(supply uint64, startHeight, endHeight uint32) uint64 {
	for i := startHeight; i < endHeight; i++ {
		supply += BlockRewardBySupply(supply, i)
	}
	return supply
}

const TotalSupply = uint64(2_100_000_000_000_000)

const InitialSupply = uint64(252_000_000_000_000)

const EmissionTailStart = uint32(48_692_960)

const EmissionTailReward = uint64(4000)

const EmissionSpeed = uint64(4_194_304)

func BlockRewardBySupply(supply uint64, height uint32) uint64 {
	if height == 0 {
		return 0
	}
	remaining := TotalSupply - supply
	if height >= EmissionTailStart && remaining >= EmissionTailReward {
		return EmissionTailReward
	}
	remainder := remaining % EmissionSpeed
	return (remaining - remainder) / EmissionSpeed
}
