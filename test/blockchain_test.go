package main

import (
	"fmt"
	"testing"

	"github.com/JI-0/private-cryptocurrency/blockchain"
)

// Test creation of chain
func TestCreationOfChainAndBlocks(t *testing.T) {
	chain := blockchain.NewChain()
	chain.AddBlock("data0")
	chain.AddBlock("data1")
	chain.AddBlock("data2")

	for _, block := range chain.Blocks {
		fmt.Printf("Prev hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		// if block.Data == []byte("genesis") {

		// }
	}
}
