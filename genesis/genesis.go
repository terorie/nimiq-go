// Package genesis provides Nimiq network genesis configs.
package genesis

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/pelletier/go-toml"
	"terorie.dev/nimiq/accounts"
	"terorie.dev/nimiq/beserial"
	"terorie.dev/nimiq/wire"
)

// Profile describes genesis config and related blockchain data.
type Profile struct {
	Config   *Config
	Block    wire.Block
	Accounts []Account
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

// InitAccounts inserts the genesis accounts to the accounts tree.
func (i *Profile) InitAccounts(a *accounts.Accounts) error {
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

//go:embed files/*
var embedFiles embed.FS

// Hardcoded profile IDs for use in OpenProfile.
const (
	ProfileMain = "main"
	ProfileTest = "test"
)

// OpenProfile reads a genesis profile from hardcoded info or the file system.
func OpenProfile(path string) (*Profile, error) {
	profiles, err := fs.Sub(embedFiles, "files")
	if err != nil {
		panic("invalid hardcoded profiles: " + err.Error())
	}
	switch path {
	case ProfileMain:
		return openProfileFromFS(profiles, ProfileMain)
	case ProfileTest:
		return openProfileFromFS(profiles, ProfileTest)
	}
	return openProfileFromFS(os.DirFS("."), path)
}

func openProfileFromFS(files fs.FS, path string) (*Profile, error) {
	tomlPath := path + ".toml"
	conf, err := readConfigFromFS(files, tomlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config toml: %w", err)
	}
	info := new(Profile)
	info.Config = conf
	blockPath := path + ".block.bin"
	blockBuf, err := fs.ReadFile(files, blockPath)
	if err != nil {
		return nil, fmt.Errorf(`failed to read block: %w`, err)
	}
	if err := beserial.UnmarshalFull(blockBuf, &info.Block); err != nil {
		return nil, fmt.Errorf(`failed to unmarshal block: %w`, err)
	}
	accountsPath := path + ".accounts.bin"
	accountsBuf, err := fs.ReadFile(files, accountsPath)
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

func readConfigFromFS(files fs.FS, path string) (conf *Config, err error) {
	f, openErr := files.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	defer f.Close()
	dec := toml.NewDecoder(f)
	conf = new(Config)
	err = dec.Decode(conf)
	return
}
