package accounts

import (
	"testing"

	"github.com/stretchr/testify/require"
	"terorie.dev/nimiq/tree"
	"terorie.dev/nimiq/wallet"
	"terorie.dev/nimiq/wire"
)

func TestAccounts_Pruning(t *testing.T) {
	accounts := NewAccounts(&tree.PMTree{Store: tree.NewMemStore()})
	w := wallet.GenerateBasic()

	// Give a block reward.
	require.NoError(t, accounts.Push(&wire.Block{
		Header: wire.BlockHeader{
			Height: 1,
		},
		Body: &wire.BlockBody{
			MinerAddr: w.GetAddress(),
		},
	}))

	// Create vesting contract.
	createTx := wire.ExtendedTx{
		Sender:              w.GetAddress(),
		SenderType:          wire.AccountBasic,
		RecipientType:       wire.AccountVesting,
		Value:               100,
		Fee:                 0,
		ValidityStartHeight: 1,
		Flags:               wire.TxFlagContractCreation,
		Data: []byte{
			// Address
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			// Vesting step blocks
			0x00, 0x00, 0x00, 0x01,
		},
		NetworkID: 4,
	}
	w.SignExtendedTx(&createTx)
	require.NoError(t, accounts.Push(&wire.Block{
		Header: wire.BlockHeader{
			Height: 2,
		},
		Body: &wire.BlockBody{
			MinerAddr: w.GetAddress(),
			ExtraData: nil,
			Txs: []wire.WrapTx{
				{Tx: &createTx},
			},
		},
	}))

	// TODO Empty vesting contract without pruning it.
}
