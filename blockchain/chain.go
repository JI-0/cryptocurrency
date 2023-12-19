package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "first transaction from genesis"
)

type Chain struct {
	LastHash []byte
	Database *badger.DB
}

type ChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func Genesis(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewChain(address string) *Chain {
	if DBExists() {
		fmt.Println("Chain already exists")
		runtime.Goexit()
	}

	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("Creating new blockchain")
			cbtx := CoinbaseTransaction(address, genesisData)
			genesis := Genesis(cbtx)
			if err = txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
				return err
			}
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else if err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}

	blockchain := Chain{lastHash, db}
	return &blockchain
}

func ContinueChain(address string) *Chain {
	if DBExists() == false {
		fmt.Println("No chain exists")
		runtime.Goexit()
	}

	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		if item, err := txn.Get([]byte("lh")); err != nil {
			return err
		} else {
			item.Value(func(val []byte) error {
				lastHash = val
				return nil
			})
		}
		return nil
	}); err != nil {
		panic(err)
	}

	blockchain := Chain{lastHash, db}
	return &blockchain
}

func (c *Chain) AddBlock(transactions []*Transaction) {
	var lastHash []byte
	if err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		return nil
	}); err != nil {
		panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	if err := c.Database.Update(func(txn *badger.Txn) error {
		if err := txn.Set(newBlock.Hash, newBlock.Serialize()); err != nil {
			return err
		}
		if err := txn.Set([]byte("lh"), newBlock.Hash); err != nil {
			return err
		}
		c.LastHash = newBlock.Hash
		return nil
	}); err != nil {
		panic(err)
	}
}

func (c *Chain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	previousTxs := make(map[string]Transaction)
	for _, input := range tx.Inputs {
		previousTx, err := c.FindTransaction(input.ID)
		if err != nil {
			panic(err)
		}
		previousTxs[hex.EncodeToString(previousTx.ID)] = previousTx
	}
	tx.Sign(privateKey, previousTxs)
}

func (c *Chain) VerifyTransaction(tx *Transaction) bool {
	previousTxs := make(map[string]Transaction)
	for _, input := range tx.Inputs {
		previousTx, err := c.FindTransaction(input.ID)
		if err != nil {
			panic(err)
		}
		previousTxs[hex.EncodeToString(previousTx.ID)] = previousTx
	}
	return tx.Verify(previousTxs)
}

func (c *Chain) FindTransaction(ID []byte) (Transaction, error) {
	iter := c.Iterator()
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("transaction does not exist")
}

func (c *Chain) FindUnspentTransactions(publicKeyHash []byte) []Transaction {
	var unspentTxs []Transaction
	spentTxs := make(map[string][]int)
	iter := c.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTxs[txID] != nil {
					for _, spentOut := range spentTxs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(publicKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbaseTransaction() == false {
				for _, in := range tx.Inputs {
					if in.UsesKey(publicKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spentTxs[inTxID] = append(spentTxs[inTxID], in.Output)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

func (c *Chain) FindUTXO(publicKeyHash []byte) []TransactionOutput {
	var UTXOs []TransactionOutput
	unspentTransactions := c.FindUnspentTransactions(publicKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(publicKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (c *Chain) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := c.FindUnspentTransactions(publicKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func (c *Chain) Iterator() *ChainIterator {
	return &ChainIterator{c.LastHash, c.Database}
}

// Iterate backwards
func (it *ChainIterator) Next() *Block {
	var block *Block
	if err := it.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(it.CurrentHash)
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			block = block.Deserialize(val)
			return nil
		})
		return nil
	}); err != nil {
		panic(err)
	}
	it.CurrentHash = block.PrevHash
	return block
}
