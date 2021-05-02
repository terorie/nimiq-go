package main

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"terorie.dev/nimiq/beserial"
	"terorie.dev/nimiq/wire"
)

func main() {
	// untar()
	loadblock()
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
	}
}
