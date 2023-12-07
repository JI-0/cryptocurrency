package main

import (
	"bytes"
	"crypto/sha512"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

func NewBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash}
	block.deriveHash()
	return block
}

func (b *Block) deriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha512.Sum512(info)
	b.Hash = hash[:]
}
