package main

import (
	"archive/tar"
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"terorie.dev/nimiq/accounts"
	"terorie.dev/nimiq/beserial"
	"terorie.dev/nimiq/genesis"
	"terorie.dev/nimiq/tree"
	"terorie.dev/nimiq/wire"
)

func main() {
	blocksPath := flag.String("blocksPath", "", "Blocks dump file (required)")
	profile := flag.String("profile", genesis.ProfileTest, "Genesis profile")
	debug := flag.Bool("debug", false, "Print debug information")
	flag.Parse()

	if *blocksPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(*blocksPath)
	if err != nil {
		panic("failed to load blocks: " + err.Error())
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	archive := tar.NewReader(rd)

	store := tree.NewMemStore()
	pmTree := tree.PMTree{Store: store}
	accs := accounts.NewAccounts(&pmTree)

	inf, err := genesis.OpenProfile(*profile)
	if err != nil {
		panic("failed to load profile: " + err.Error())
	}
	if err := inf.InitAccounts(accs); err != nil {
		panic(err.Error())
	}

	start := time.Now()
	blocks := 0
	txs := 0
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
		if *debug {
			fmt.Printf("Block % 6d\n", block.Header.Height)
		}
		if block.Header.AccountsHash != hash {
			panic("accounts hash mismatch, expected: " + hex.EncodeToString(block.Header.AccountsHash[:]))
		}
		blocks++
		txs += len(block.Body.Txs)
	}
	fmt.Println("Blocks:", blocks)
	fmt.Println("Txs:", txs)
	fmt.Println("Time:", time.Since(start))
}
