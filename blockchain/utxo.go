package blockchain

import (
	"bytes"
	"encoding/hex"

	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix = []byte("utxo-")
	prefixLen  = len(utxoPrefix)
)

type UTXOSet struct {
	Chain *Chain
}

func (u UTXOSet) Reindex() {
	db := u.Chain.Database
	u.DeleteByPrefix(utxoPrefix)
	UTXOs := u.Chain.FindUTXOs()
	if err := db.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXOs {
			key, err := hex.DecodeString(txId)
			if err != nil {
				panic(err)
			}
			key = append(utxoPrefix, key...)
			err = txn.Set(key, outs.Serialize())
			if err != nil {
				panic(err)
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func (u UTXOSet) CountTransactions() int {
	db := u.Chain.Database
	counter := 0
	if err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return counter
}

func (u UTXOSet) FindUTXO(publicKeyHash []byte) []TransactionOutput {
	var UTXOs []TransactionOutput
	db := u.Chain.Database

	if err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			var outs TransactionOutputs
			item.Value(func(val []byte) error {
				outs = DeserializeOutputs(val)
				return nil
			})
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}

	return UTXOs
}

func (u UTXOSet) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := u.Chain.Database

	if err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			var outs TransactionOutputs
			item.Value(func(val []byte) error {
				outs = DeserializeOutputs(val)
				return nil
			})
			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}

	return accumulated, unspentOuts
}

func (u *UTXOSet) Update(b *Block) {
	db := u.Chain.Database
	if err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range b.Transactions {
			if !tx.IsCoinbaseTransaction() {
				for _, in := range tx.Inputs {
					updatedOuts := TransactionOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					if err != nil {
						panic(err)
					}
					var outs TransactionOutputs
					item.Value(func(val []byte) error {
						outs = DeserializeOutputs(val)
						return nil
					})
					for outId, out := range outs.Outputs {
						if outId != in.Output {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}
					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							panic(err)
						}
					} else {
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							panic(err)
						}
					}
				}
			}
			newOutputs := TransactionOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				panic(err)
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelection [][]byte) error {
		if err := u.Chain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelection {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	collectionSize := 100000
	u.Chain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelection := make([][]byte, 0, collectionSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelection = append(keysForDelection, key)
			keysCollected++
			if keysCollected == collectionSize {
				if err := deleteKeys(keysForDelection); err != nil {
					panic(err)
				}
				keysForDelection = make([][]byte, 0, collectionSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelection); err != nil {
				panic(err)
			}
		}
		return nil
	})
}
