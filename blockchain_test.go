package main

import (
	"fmt"
	"testing"
)

// Test creation of chain
func TestCreationOfChainAndBlocks(t *testing.T) {
	chain := NewChain()
	chain.addBlock("data0")
	chain.addBlock("data1")
	chain.addBlock("data2")

	for _, block := range chain.blocks {
		fmt.Printf("Prev hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		// if block.Data == []byte("genesis") {

		// }
	}
}
