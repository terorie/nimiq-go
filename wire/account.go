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
	AccountUnknown = 0xFF
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

// An OverspendError means an attempt was made to spend more funds than available.
type OverspendError struct {
	Available, Spend uint64
}

func (o *OverspendError) Error() string {
	return fmt.Sprintf("trying to spend %d but only has %d", o.Spend, o.Available)
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
	return c.Value == 0
}

func (c *VestingAccount) Init(tx Tx, _ uint32, _ uint64) error {
	ext := tx.(*ExtendedTx)
	var err error
	switch len(ext.Data) {
	case 24:
		var params struct {
			Address           [20]byte
			VestingStepBlocks uint32
		}
		err = beserial.UnmarshalFull(ext.Data, &params)
		if err != nil {
			return err
		}
		c.Owner = params.Address
		c.VestingStart = 0
		c.VestingStepBlocks = params.VestingStepBlocks
		c.VestingStepAmount = ext.Value
		c.VestingTotalAmount = ext.Value
	case 36:
		var params struct {
			Address           [20]byte
			VestingStart      uint32
			VestingStepBlocks uint32
			VestingStepAmount uint64
		}
		err = beserial.UnmarshalFull(ext.Data, &params)
		if err != nil {
			return err
		}
		c.Owner = params.Address
		c.VestingStart = params.VestingStart
		c.VestingStepBlocks = params.VestingStepBlocks
		c.VestingStepAmount = params.VestingStepAmount
		c.VestingTotalAmount = ext.Value
	case 44:
		var params struct {
			Address            [20]byte
			VestingStart       uint32
			VestingStepBlocks  uint32
			VestingStepAmount  uint64
			VestingTotalAmount uint64
		}
		err = beserial.UnmarshalFull(ext.Data, &params)
		if err != nil {
			return err
		}
		c.Owner = params.Address
		c.VestingStart = params.VestingStart
		c.VestingStepBlocks = params.VestingStepBlocks
		c.VestingStepAmount = params.VestingStepAmount
		c.VestingTotalAmount = params.VestingTotalAmount
	default:
		return fmt.Errorf("invalid vesting account")
	}
	return nil
}

// ApplyOutgoingTx removes the transaction value and fee from the account.
func (c *VestingAccount) ApplyOutgoingTx(tx Tx, height uint32) (Account, error) {
	unlocked := c.amountUnlocked(height)
	// Check transaction amount.
	extTx := tx.(*ExtendedTx)
	if extTx.Value > unlocked {
		return nil, &OverspendError{Available: unlocked, Spend: extTx.Value}
	}
	newValue := c.Value - extTx.Value
	// Verify transaction was signed by contract owner.
	var proof SignatureProof
	if err := beserial.UnmarshalFull(extTx.Proof, &proof); err != nil {
		return nil, fmt.Errorf("in wire.SignatureProof: %w", err)
	}
	if proof.SignerAddress() != c.Owner {
		return nil, fmt.Errorf("invalid signature") // TODO use const error
	}
	// Create copy.
	updated := new(VestingAccount)
	*updated = *c
	updated.Value = newValue
	return updated, nil
}

// ApplyIncomingTx adds the transaction value to the account.
func (c *VestingAccount) ApplyIncomingTx(tx Tx, _ uint32) (Account, error) {
	panic("not implemented")
}

func (c *VestingAccount) amountUnlocked(height uint32) uint64 {
	if height <= c.VestingStart {
		return 0 // cannot spend funds before vesting start
	}
	if c.VestingStepBlocks == 0 || c.VestingStepAmount == 0 {
		return c.Value // unlock immediately
	}
	// unlock step by step
	progress := float64(height-c.VestingStart) / float64(c.VestingStepBlocks)
	unlocked := uint64(progress * float64(c.VestingStepAmount))
	return unlocked
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
	// TODO use typed errors
	ext := tx.(*ExtendedTx)
	var params struct {
		Sender    [20]byte
		Recipient [20]byte
		Hash      Hash
		HashCount uint8
		Timeout   uint32
	}
	if err := beserial.UnmarshalFull(ext.Data, &params); err != nil {
		return fmt.Errorf("invalid HTLC creation data: %w", err)
	}
	if params.HashCount == 0 {
		return fmt.Errorf("invalid hash count: %d", params.HashCount)
	}
	// TODO Argon2d type allowed?
	*h = HTLCAccount{
		Value:       prevBalance,
		Sender:      params.Sender,
		Recipient:   params.Recipient,
		Hash:        params.Hash,
		HashCount:   params.HashCount,
		Timeout:     params.Timeout,
		TotalAmount: ext.Value,
	}
	return nil
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
