// Package accounts implements the Nimiq accounts tree.
package accounts

import (
	"encoding/hex"
	"fmt"

	"terorie.dev/nimiq/policy"
	"terorie.dev/nimiq/tree"
	"terorie.dev/nimiq/wire"
)

// Accounts is the main interface to the accounts trie.
// It contains all accounts and contracts at the blockchain head.
//
// It consists of Ethereum's Merkle Patricia Trie Specification.
// The 32-byte root of the state is saved in each block header to allow for easy verification.
// "Merkle paths" can be used to create cryptographic proofs of a particular (historical) state entry,
// e.g. the account balance at a certain block height.
//
// From a high-level perspective, state updates are applied by committing or reverting a block.
// External applications can only probe the state just before and after committing any block,
// as updates are applied atomically, in a single (database) transaction.
// Internally, while committing a block, the state undergoes a series of smaller transient steps.
// That means re-ordering of (Nimiq) transactions in a block could theoretically lead to different state outcomes.
type Accounts struct {
	Tree *tree.PMTree
}

// NewAccounts creates a new Accounts trie interface backed by the specified store.
func NewAccounts(tree *tree.PMTree) *Accounts {
	return &Accounts{
		Tree: tree,
	}
}

// Push pushes a new block to the state.
func (a *Accounts) Push(block *wire.Block) error {
	if err := a.pushSenders(block, func(acc wire.Account, tx wire.Tx, height uint32) (wire.Account, error) {
		return acc.ApplyOutgoingTx(tx, height)
	}); err != nil {
		return err
	}
	if err := a.pushRecipients(block, func(acc wire.Account, tx wire.Tx, height uint32) (wire.Account, error) {
		return acc.ApplyIncomingTx(tx, height)
	}); err != nil {
		return err
	}
	if err := a.createContracts(block); err != nil {
		return err
	}
	if err := a.pruneAccounts(block); err != nil {
		return err
	}
	if err := a.pushInherents(block); err != nil {
		return err
	}
	return nil
}

// Revert undoes all changes a block does to the state.
// It is has the exact inverse effects of Push.
func (a *Accounts) Revert(block *wire.Block) error {
	panic("not implemented")
}

func (a *Accounts) GetAccount(addr *[20]byte) wire.Account {
	buf := a.Tree.GetEntry(addr)
	if len(buf) <= 0 {
		return &wire.InitialAccount
	}
	var acc wire.WrapAccount
	n, err := acc.UnmarshalBESerial(buf)
	if err != nil {
		panic(fmt.Sprintf("failed to read account %s: %s",
			hex.EncodeToString(addr[:]), err))
	}
	if n != len(buf) {
		panic(fmt.Sprintf("did not fully read account %s: %d vs %d bytes",
			hex.EncodeToString(addr[:]), n, len(buf)))
	}
	return acc.Account
}

func (a *Accounts) PutAccount(addr *[20]byte, acc wire.Account) {
	wrap := wire.WrapAccount{Account: acc}
	buf, err := wrap.MarshalBESerial(nil)
	if err != nil {
		panic(err)
	}
	a.Tree.PutEntry(addr, buf)
}

func (a *Accounts) pushSenders(block *wire.Block, op AccountOp) error {
	for _, tx := range block.Body.Txs {
		content := tx.Tx.AsTxContent()
		if err := a.pushTx(&content.Sender, content.SenderType, tx.Tx, block.Header.Height, op); err != nil {
			return err
		}
	}
	return nil
}

func (a *Accounts) pushRecipients(block *wire.Block, op AccountOp) error {
	for _, tx := range block.Body.Txs {
		content := tx.Tx.AsTxContent()
		var recipientType uint8
		if content.Flags&wire.TxFlagContractCreation == 0 {
			recipientType = content.RecipientType
		}
		if err := a.pushTx(&content.Recipient, recipientType, tx.Tx, block.Header.Height, op); err != nil {
			return err
		}
	}
	return nil
}

func (a *Accounts) createContracts(block *wire.Block) error {
	for _, tx := range block.Body.Txs {
		content := tx.Tx.AsTxContent()
		if content.Flags&wire.TxFlagContractCreation != 0 {
			if err := a.createContract(tx.Tx, block.Header.Height); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Accounts) pushInherents(block *wire.Block) error {
	var reward uint64
	reward = policy.BlockReward(block.Header.Height)
	for _, tx := range block.Body.Txs {
		_, fee := tx.Tx.Amount()
		reward += fee
	}
	coinbase := wire.BasicTx{
		Recipient:           block.Body.MinerAddr,
		Value:               reward,
		ValidityStartHeight: block.Header.Height,
	}
	return a.pushTx(&block.Body.MinerAddr, wire.AccountBasic, &coinbase, block.Header.Height,
		func(acc wire.Account, tx wire.Tx, height uint32) (wire.Account, error) {
			return acc.ApplyIncomingTx(tx, height)
		})
}

// pruneAccounts removes accounts from the state.
// It compares the actual accounts to prune with the ones in the block.
// An error is returned if
//  - a non-empty account is pruned
//  - an empty account is not pruned
//  - the pruned account entry is different from the actual account
func (a *Accounts) pruneAccounts(block *wire.Block) error {
	// Check whether each account in the pruned list is indeed to be pruned,
	// if they were correctly pruned and if accounts were missed.
	pruned := make(map[[20]byte]wire.Account)
	for _, acc := range block.Body.Pruned {
		pruned[acc.Address] = acc.Account
	}
	// For each transaction that prunes an account,
	// check whether the prune was correctly quoted in the prune section of the block.
	for _, tx := range block.Body.Txs {
		senderAddr, _ := tx.Tx.GetSender()
		sender := a.GetAccount(senderAddr)
		if !sender.IsEmpty() {
			continue
		}
		// We found a transaction that forces removal of the sender account.
		prunedAcc, ok := pruned[*senderAddr]
		if !ok {
			return &InvalidPruneError{
				address: *senderAddr,
				error:   fmt.Errorf("account missing in prune list"),
			}
		}
		if !wire.AccountEq(prunedAcc, sender) {
			return &InvalidPruneError{
				address: *senderAddr,
				error:   fmt.Errorf("altered account in prune list"),
			}
		}
		// Delete account from state tree.
		a.Tree.PutEntry(senderAddr, nil)
		delete(pruned, *senderAddr)
	}
	// Check if the block specified prunes without justification.
	if len(pruned) > 0 {
		// Get random address of an illegally-pruned account.
		var address [20]byte
		for a := range pruned {
			address = a
			break
		}
		return &InvalidPruneError{
			address: address,
			error:   fmt.Errorf("account pruned too early"),
		}
	}
	return nil
}

func (a *Accounts) createContract(tx wire.Tx, height uint32) error {
	// TODO Cache this (WrapTx)
	content := tx.AsTxContent()
	if content.Flags&wire.TxFlagContractCreation == 0 {
		panic("accounts: createContract called on tx without contract creation flag")
	}
	acc := a.GetAccount(&content.Recipient)
	var newAcc wire.Account
	switch content.RecipientType {
	case wire.AccountBasic:
		return &InvalidForRecipientError{
			address: content.Recipient,
			error:   fmt.Errorf("contract creation for basic account"),
		}
	case wire.AccountVesting:
		newAcc = new(wire.VestingAccount)
	case wire.AccountHTLC:
		newAcc = new(wire.HTLCAccount)
	default:
		return &InvalidForRecipientError{
			address: content.Recipient,
			error:   fmt.Errorf("invalid recipient type: %d", content.RecipientType),
		}
	}
	return newAcc.Init(tx, height, acc.Balance())
}

func (a *Accounts) pushTx(address *[20]byte, accType uint8, tx wire.Tx, height uint32, op AccountOp) error {
	account := a.GetAccount(address)
	if accType != account.Type() {
		return &AccountTypeMismatchError{
			Expected: accType,
			Got:      account.Type(),
		}
	}
	newAccount, err := op(account, tx, height)
	if err != nil {
		return err
	}
	a.PutAccount(address, newAccount)
	return nil
}

type AccountOp func(acc wire.Account, tx wire.Tx, height uint32) (wire.Account, error)

// AccountTypeMismatchError gets raised when the expected account type
// does not match with the one specified a transaction.
type AccountTypeMismatchError struct {
	Expected, Got uint8
}

func (a *AccountTypeMismatchError) Error() string {
	return fmt.Sprintf("account type mismatch: got %d want %d",
		a.Got, a.Expected)
}

type InvalidForSenderError struct {
	address [20]byte
	error   error
}

func (i *InvalidForSenderError) Error() string {
	return fmt.Sprintf("invalid for sender %x: %s", &i.address, i.error)
}

func (i *InvalidForSenderError) Unwrap() error {
	return i.error
}

type InvalidForRecipientError struct {
	address [20]byte
	error   error
}

func (i *InvalidForRecipientError) Error() string {
	return fmt.Sprintf("invalid for recipient %x: %s", &i.address, i.error)
}

func (i *InvalidForRecipientError) Unwrap() error {
	return i.error
}

type InvalidPruneError struct {
	address [20]byte
	error   error
}

func (i *InvalidPruneError) Error() string {
	return fmt.Sprintf("invalid prune of account %x: %s", &i.address, i.error)
}

func (i *InvalidPruneError) Unwrap() error {
	return i.error
}
