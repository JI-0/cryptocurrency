package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/JI-0/private-cryptocurrency/blockchain"
)

const dbPath = "./tmp/blocks"

// Test creation of chain and POW
func TestCreationOfChainAndBlocks(t *testing.T) {
	//Delete files from previous test
	if err := os.RemoveAll(dbPath); err != nil {
		t.Fatal("Database file error: ", err)
	}
	//Start test
	chain := blockchain.NewChain("")
	// chain.AddBlock("data0")
	// chain.AddBlock("data1")
	// chain.AddBlock("data2")

	iterator := chain.Iterator()
	for {
		block := iterator.Next()

		fmt.Printf("Prev hash: %x\n", block.PrevHash)
		// fmt.Printf("Data: %s\n", block.Transactions)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		if !pow.Validate() {
			t.Fatalf(`Proof of work returned invalid`)
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
}
