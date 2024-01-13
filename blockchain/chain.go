package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks_%s"
)

type Chain struct {
	LastHash []byte
	Database *badger.DB
}

func Genesis(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func DBExists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

func DBRetry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`Removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func DBOpen(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := DBRetry(dir, opts); err == nil {
				println("Database unlocked, value log truncated")
				return db, nil
			}
			fmt.Println("Could not unlock database: ", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}

func NewChain(address, path string) *Chain {
	path = fmt.Sprintf(dbPath, path)
	if DBExists(path) {
		fmt.Println("Chain already exists")
		runtime.Goexit()
	}

	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)

	db, err := DBOpen(path, opts)
	if err != nil {
		panic(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("Creating new blockchain")
			cbtx := CoinbaseTransaction(address, "coinbase")
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

func ContinueChain(path string) *Chain {
	path = fmt.Sprintf(dbPath, path)
	if DBExists(path) == false {
		fmt.Println("No chain exists")
		runtime.Goexit()
	}

	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)

	db, err := DBOpen(path, opts)
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

func (c *Chain) AddBlock(block *Block) {
	if err := c.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}
		blockData := block.Serialize()
		if err := txn.Set(block.Hash, blockData); err != nil {
			return err
		}
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}
		var lastHash []byte
		item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		item, err = txn.Get(lastHash)
		if err != nil {
			return err
		}
		var lastBlock *Block
		if err := item.Value(func(val []byte) error {
			lastBlock = lastBlock.Deserialize(val)
			return nil
		}); err != nil {
			return err
		}

		if block.Height > lastBlock.Height {
			if err := txn.Set([]byte("lh"), block.Hash); err != nil {
				return err
			}
			c.LastHash = block.Hash
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func (c *Chain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	if err := c.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return err
		} else {
			item.Value(func(val []byte) error {
				block = *block.Deserialize(val)
				return nil
			})
		}
		return nil
	}); err != nil {
		return block, err
	}
	return block, nil
}

func (c *Chain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := c.Iterator()
	for {
		block := iter.Next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return blocks
}

func (c *Chain) GetTopHeight() int {
	var lastBlock Block
	if err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			panic(err)
		}
		var lastHash []byte
		item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		item, err = txn.Get(lastHash)
		item.Value(func(val []byte) error {
			lastBlock = *lastBlock.Deserialize(val)
			return nil
		})
		return nil
	}); err != nil {
		panic(err)
	}
	return lastBlock.Height
}

func (c *Chain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int
	if err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		var lastBlock *Block
		item, err = txn.Get(lastHash)
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			lastBlock = lastBlock.Deserialize(val)
			lastHeight = lastBlock.Height
			return nil
		})
		return nil
	}); err != nil {
		panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

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

	return newBlock
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

func (c *Chain) FindUTXOs() map[string]TransactionOutputs {
	UTXOs := make(map[string]TransactionOutputs)
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
				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXOs[txID] = outs
			}
			if tx.IsCoinbaseTransaction() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTxs[inTxID] = append(spentTxs[inTxID], in.Output)
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXOs
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
	//Verification due to mining
	if tx.IsCoinbaseTransaction() {
		return true
	}

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

func (c *Chain) Iterator() *ChainIterator {
	return &ChainIterator{c.LastHash, c.Database}
}
