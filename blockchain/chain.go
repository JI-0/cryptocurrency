package blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

const dbPath = "./tmp/blocks"

type Chain struct {
	LastHash []byte
	Database *badger.DB
}

type ChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func Genesis() *Block {
	return NewBlock("genesis", []byte{})
}

func NewChain() *Chain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		if item, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("Creating new blockchain")
			genesis := Genesis()
			if err = txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
				return err
			}
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else if err != nil {
			return err
		} else {
			item.Value(func(val []byte) error {
				lastHash = val
				return nil
			})
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}

	blockchain := Chain{lastHash, db}
	return &blockchain
}

func (c *Chain) AddBlock(data string) {
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
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash)

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
		log.Panic(err)
	}
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
		log.Panic(err)
	}
	it.CurrentHash = block.PrevHash
	return block
}
