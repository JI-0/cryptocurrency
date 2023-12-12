package test

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
	//Create chain
	chain := blockchain.NewChain("Tester-0")
	defer chain.Database.Close()
	//Get balance
	balance := 0
	UTXOs := chain.FindUTXO("Tester-0")
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Send amount 20
	tx := blockchain.NewTransaction("Tester-0", "Tester-1", 20, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Sent amount 20")
	//Get balances
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO("Tester-0")
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO("Tester-1")
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-1", balance)
	//Send amount 80
	tx = blockchain.NewTransaction("Tester-0", "Tester-1", 80, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Sent amount 80")
	//Get balances
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO("Tester-0")
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO("Tester-1")
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-1", balance)

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
