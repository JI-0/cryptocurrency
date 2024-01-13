package blockchain

import "github.com/dgraph-io/badger"

type ChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
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
