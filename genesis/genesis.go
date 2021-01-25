// Package genesis contains several hardcoded network genesis blocks to choose from.
package genesis

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"terorie.dev/nimiq/accounts"
	"terorie.dev/nimiq/beserial"
	"terorie.dev/nimiq/wire"
)

// InitInfo describes genesis config and related blockchain data.
type InitInfo struct {
	Config   *Config
	Block    wire.Block
	Accounts []Account `beserial:"len_tag=uint16"`
}

type Context struct {
	Config *Config
	Header wire.BlockHeader
}

// Account is a genesis account entry.
type Account struct {
	Address [20]byte
	Account wire.WrapAccount
}

// Config describes a genesis configuration,
// including ID, name and genesis block.
type Config struct {
	NetworkID   uint8
	Name        string
	SeedPeers   []string
	SeedLists   []string
	GenesisHash [32]byte
}

func ReadInfo(path string) (*InitInfo, error) {
	ext := filepath.Ext(path)
	if ext != ".toml" {
		return nil, fmt.Errorf("not a toml file: %s", path)
	}
	conf, err := ReadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config toml: %w", err)
	}
	info := new(InitInfo)
	info.Config = conf
	blockPath := path[:len(path)-len(ext)] + ".block.bin"
	blockBuf, err := ioutil.ReadFile(blockPath)
	if err != nil {
		return nil, fmt.Errorf(`failed to read block: %w`, err)
	}
	if err := beserial.UnmarshalFull(blockBuf, &info.Block); err != nil {
		return nil, fmt.Errorf(`failed to unmarshal block: %w`, err)
	}
	accountsPath := path[:len(path)-len(ext)] + ".accounts.bin"
	accountsBuf, err := ioutil.ReadFile(accountsPath)
	if err != nil {
		return nil, fmt.Errorf(`failed to read accounts: %w`, err)
	}
	var accountsList struct {
		Accounts []Account `beserial:"len_tag=uint16"`
	}
	if err := beserial.UnmarshalFull(accountsBuf, &accountsList); err != nil {
		return nil, fmt.Errorf(`failed to unmarshal accounts: %w`, err)
	}
	info.Accounts = accountsList.Accounts
	return info, nil
}

func ReadConfig(path string) (conf *Config, err error) {
	f, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	defer f.Close()
	dec := toml.NewDecoder(f)
	conf = new(Config)
	err = dec.Decode(conf)
	return
}

// InitAccounts inserts the genesis accounts to the accounts tree.
func (i *InitInfo) InitAccounts(a *accounts.Accounts) error {
	for _, acc := range i.Accounts {
		a.PutAccount(&acc.Address, acc.Account.Account)
	}
	if err := a.Push(&i.Block); err != nil {
		return fmt.Errorf("failed to push genesis block: %w", err)
	}
	if hash := a.Tree.Hash(); hash != i.Block.Header.AccountsHash {
		return fmt.Errorf("unexpected tree hash: %x vs %x",
			hash, i.Block.Header.AccountsHash)
	}
	return nil
}
