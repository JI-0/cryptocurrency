package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/JI-0/private-cryptocurrency/blockchain"
	"github.com/JI-0/private-cryptocurrency/wallet"
)

const (
	dbPath = "./tmp/blocks_test"
)

// Test creation of chain and POW
func TestCreationOfChainAndBlocks(t *testing.T) {
	//Delete files from previous test
	os.RemoveAll("./tmp")
	os.Mkdir("./tmp/", 0700)
	os.Mkdir("./tmp/wallets", 0700)
	if err := os.RemoveAll(dbPath); err != nil {
		t.Fatal("Database file error: ", err)
	}
	os.Mkdir(dbPath, 0700)
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
	chain := blockchain.NewChain(string(w0), "test")
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Chain: chain}
	UTXOSet.Reindex()
	//Get balance
	balance := 0
	UTXOs := UTXOSet.FindUTXO(w0publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Send amount 20
	w0w := wallets.GetWallet(w0)
	tx := blockchain.NewTransaction(&w0w, w1, 20, &UTXOSet)
	block := chain.MineBlock([]*blockchain.Transaction{tx})
	UTXOSet.Update(block)
	fmt.Println("Sent amount 20")
	//Get balances
	//Get balance
	balance = 0
	UTXOs = UTXOSet.FindUTXO(w0publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Get balance
	balance = 0
	UTXOs = UTXOSet.FindUTXO(w1publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-1", balance)
	//Send amount 80
	tx = blockchain.NewTransaction(&w0w, w1, 80, &UTXOSet)
	cbTx := blockchain.CoinbaseTransaction(w0, "")
	cbTx0 := blockchain.CoinbaseTransaction(w1, "")
	block = chain.MineBlock([]*blockchain.Transaction{tx, cbTx, cbTx0})
	UTXOSet.Update(block)
	fmt.Println("Sent amount 80")
	//Get balances
	//Get balance
	balance = 0
	UTXOs = UTXOSet.FindUTXO(w0publicKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", "Tester-0", balance)
	//Get balance
	balance = 0
	UTXOs = UTXOSet.FindUTXO(w1publicKeyHash)
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

		pow := blockchain.NewProof(chain, block, true)
		if !pow.Validate() {
			t.Fatalf(`Proof of work returned invalid`)
		}
		pow.Destroy()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}
