package wire

import (
	"fmt"
	"strconv"

	"golang.org/x/crypto/blake2b"
	"terorie.dev/nimiq/beserial"
)

// Account is implemented by all accounts types.
type Account interface {
	Balance() uint64
	Type() uint8
	IsEmpty() bool // empty accounts get pruned
	// Init creates the account from a transaction
	Init(tx Tx, height uint32, prevBalance uint64) error
	ApplyOutgoingTx(tx Tx, height uint32) (Account, error)
	ApplyIncomingTx(tx Tx, height uint32) (Account, error)
}

// Account types.
const (
	AccountBasic   = 0
	AccountVesting = 1
	AccountHTLC    = 2
	AccountStaking = 3 // reserved
)

// WrapAccount implements BESerial for Account.
type WrapAccount struct {
	Account Account
}

func (wa *WrapAccount) UnmarshalBESerial(buf []byte) (n int, err error) {
	if len(buf) < 1 {
		return 0, beserial.ErrUnexpectedEOF
	}
	accType := buf[0]
	n++
	buf = buf[1:]
	switch accType {
	case AccountBasic:
		wa.Account = new(BasicAccount)
	case AccountVesting:
		wa.Account = new(VestingAccount)
	case AccountHTLC:
		wa.Account = new(HTLCAccount)
	default:
		return 0, fmt.Errorf("invalid account type: %d", accType)
	}
	var sub int
	sub, err = beserial.Unmarshal(buf, wa.Account)
	n += sub
	return
}

func (wa *WrapAccount) MarshalBESerial(b []byte) ([]byte, error) {
	b = append(b, wa.Account.Type())
	return beserial.Marshal(b, wa.Account)
}

func (wa *WrapAccount) SizeBESerial() (n int, err error) {
	n = 1
	var sub int
	sub, err = beserial.Size(wa.Account)
	n += sub
	return
}

var InitialAccount = BasicAccount{Value: 0}

// AccountEq checks two accounts for equality.
// If either argument is nil, false is returned.
func AccountEq(a1, a2 Account) bool {
	if a1 == nil || a2 == nil {
		return false
	}
	a1Type := a1.Type()
	if a1Type != a2.Type() {
		return false
	}
	switch a1Type {
	case AccountBasic:
		return *(a1.(*BasicAccount)) == *(a2.(*BasicAccount))
	case AccountVesting:
		return *(a1.(*VestingAccount)) == *(a2.(*VestingAccount))
	case AccountHTLC:
		panic("not implemented")
		//return *(a1.(*HTLCAccount)) == *(a2.(*HTLCAccount))
	default:
		panic("invalid account type: " + strconv.FormatUint(uint64(a1Type), 10))
	}
}

type OverspendError struct {
	Available, Spend uint64
}

func (o *OverspendError) Error() string {
	return fmt.Sprintf("overspending: have %d, spending %d", o.Available, o.Spend)
}

// A BasicAccount is a simple account that can hold funds (balance)
// and is controlled by an Ed25519 private key.
// https://nimiq-network.github.io/developer-reference/chapters/accounts-and-contracts.html#basic-account
type BasicAccount struct {
	Value uint64
}

func (a *BasicAccount) Balance() uint64 {
	return a.Value
}

func (a *BasicAccount) Type() uint8 {
	return AccountBasic
}

func (a *BasicAccount) IsEmpty() bool {
	return a.Value == 0
}

func (a *BasicAccount) Init(tx Tx, height uint32, prevBalance uint64) error {
	panic("not implemented")
}

// ApplyOutgoingTx removes the transaction value and fee from the account.
func (a *BasicAccount) ApplyOutgoingTx(tx Tx, _ uint32) (Account, error) {
	value, fee := tx.Amount()
	amount := value + fee
	if amount > a.Value {
		return nil, &OverspendError{Available: a.Value, Spend: amount}
	}
	return &BasicAccount{Value: a.Value - amount}, nil
}

// ApplyIncomingTx adds the transaction value to the account.
func (a *BasicAccount) ApplyIncomingTx(tx Tx, _ uint32) (Account, error) {
	value, _ := tx.Amount()
	return &BasicAccount{Value: a.Value + value}, nil
}

// A VestingAccount allows money to be spent in a scheduled way.
// https://nimiq-network.github.io/developer-reference/chapters/accounts-and-contracts.html#vesting-contract
type VestingAccount struct {
	Value              uint64
	Owner              [20]byte
	VestingStart       uint32
	VestingStepBlocks  uint32
	VestingStepAmount  uint64
	VestingTotalAmount uint64
}

func (c *VestingAccount) Balance() uint64 {
	return c.Value
}

func (c *VestingAccount) Type() uint8 {
	return AccountVesting
}

func (c *VestingAccount) IsEmpty() bool {
	panic("not implemented")
}

func (c *VestingAccount) Init(tx Tx, height uint32, prevBalance uint64) error {
	panic("not implemented")
}

// ApplyOutgoingTx removes the transaction value and fee from the account.
func (c *VestingAccount) ApplyOutgoingTx(tx Tx, _ uint32) (Account, error) {
	panic("not implemented")
}

// ApplyIncomingTx adds the transaction value to the account.
func (c *VestingAccount) ApplyIncomingTx(tx Tx, _ uint32) (Account, error) {
	panic("not implemented")
}

// An HTLCAccount (hashed time-locked contract)
// are used for conditional transfers, off-chain payment methods,
// and cross-chain atomic swaps.
// https://nimiq-network.github.io/developer-reference/chapters/accounts-and-contracts.html#hashed-time-locked-contract
type HTLCAccount struct {
	Value       uint64
	Sender      [20]byte
	Recipient   [20]byte
	Hash        Hash
	HashCount   uint8
	Timeout     uint32
	TotalAmount uint64
}

func (h *HTLCAccount) Balance() uint64 {
	return h.Value
}

func (h *HTLCAccount) Type() uint8 {
	return AccountHTLC
}

func (h *HTLCAccount) IsEmpty() bool {
	panic("not implemented")
}

func (h *HTLCAccount) Init(tx Tx, height uint32, prevBalance uint64) error {
	panic("not implemented")
}

// ApplyOutgoingTx removes the transaction value and fee from the account.
func (h *HTLCAccount) ApplyOutgoingTx(tx Tx, _ uint32) (Account, error) {
	panic("not implemented")
}

// ApplyIncomingTx adds the transaction value to the account.
func (h *HTLCAccount) ApplyIncomingTx(tx Tx, _ uint32) (Account, error) {
	panic("not implemented")
}

// AccountPruned marks the removal of an account from the state.
type AccountPruned struct {
	Address [20]byte
	Account WrapAccount
}

// PublicKeyToAddress derives the NIM address from the Ed25519 public key.
func PublicKeyToAddress(pub *[32]byte) (addr [20]byte) {
	h := blake2b.Sum256(pub[:])
	copy(addr[:], h[:20])
	return
}
