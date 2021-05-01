package main

import (
	"archive/tar"
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"terorie.dev/nimiq/accounts"
	"terorie.dev/nimiq/beserial"
	"terorie.dev/nimiq/genesis"
	"terorie.dev/nimiq/tree"
	"terorie.dev/nimiq/wire"
)

func main() {
	untar()
	// loadblock()
}

func loadblock() {
	f, err := os.Open("43962")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err.Error())
	}
	var block wire.Block
	if err := beserial.UnmarshalFull(buf, &block); err != nil {
		fmt.Println("Failed to unmarshal block: ", err.Error())
	}
}

func untar() {
	f, err := os.Open("blocks.tar")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	archive := tar.NewReader(rd)

	store := tree.NewMemStore()
	pmTree := tree.PMTree{Store: store}
	accs := accounts.NewAccounts(&pmTree)

	inf, err := genesis.ReadInfo("genesis/files/testnet.toml")
	if err != nil {
		panic(err.Error())
	}
	if err := inf.InitAccounts(accs); err != nil {
		panic(err.Error())
	}
	hash := pmTree.Hash()
	fmt.Println("Genesis,      accounts hash:", hex.EncodeToString(hash[:]))

	_, _ = archive.Next()
	for {
		_, err := archive.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err.Error())
		}
		buf, err := ioutil.ReadAll(archive)
		if err != nil {
			panic(err.Error())
		}
		var block wire.Block
		if err := beserial.UnmarshalFull(buf, &block); err != nil {
			panic("failed to unmarshal block: " + err.Error())
		}
		if err := accs.Push(&block); err != nil {
			panic(fmt.Sprintf("failed to commit block %d: %s", block.Header.Height, err.Error()))
		}
		hash := pmTree.Hash()
		fmt.Printf("Block % 6d, accounts hash: %x\n", block.Header.Height, hash)
		if block.Header.AccountsHash != hash {
			panic("accounts hash mismatch, expected: " + hex.EncodeToString(block.Header.AccountsHash[:]))
		}
	}
}
