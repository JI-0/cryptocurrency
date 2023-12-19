package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/JI-0/private-cryptocurrency/blockchain"
	"github.com/JI-0/private-cryptocurrency/wallet"
)

const dbPath = "./tmp/blocks"

// Test creation of chain and POW
func TestCreationOfChainAndBlocks(t *testing.T) {
	//Delete files from previous test
	if err := os.RemoveAll(dbPath); err != nil {
		t.Fatal("Database file error: ", err)
	}
	if err := os.RemoveAll(walletFile); err != nil {
		t.Fatal("Wallet file error: ", err)
	}
	if err := os.Mkdir(walletFile, 0700); err != nil {
		t.Fatal("Cannot create dir")
	}
	//Start test
	//Create wallets
	wallets, _ := wallet.NewWallets()
	w0 := wallets.AddWallet()
	w0publicKeyHash := wallet.Base58Decode([]byte(w0))
	w0publicKeyHash = w0publicKeyHash[1 : len(w0publicKeyHash)-wallet.ChecksumLen]
	w1 := wallets.AddWallet()
	w1publicKeyHash := wallet.Base58Decode([]byte(w1))
	w1publicKeyHash = w1publicKeyHash[1 : len(w1publicKeyHash)-wallet.ChecksumLen]
	wallets.Save()
	//Create chain
	chain := blockchain.NewChain(string(w0))
	defer chain.Database.Close()
	//Get balance
	balance := 0
	UTXOs := chain.FindUTXO(w0publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Send amount 20
	tx := blockchain.NewTransaction(w0, w1, 20, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Sent amount 20")
	//Get balances
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO(w0publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO(w1publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-1", balance)
	//Send amount 80
	tx = blockchain.NewTransaction(w0, w1, 80, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Sent amount 80")
	//Get balances
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO(w0publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Get balance
	balance = 0
	UTXOs = chain.FindUTXO(w1publicKeyHash)
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
