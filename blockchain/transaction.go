package blockchain

import (
	"bytes"
	"crypto/sha512"
	"encoding/gob"
	"encoding/hex"
	"fmt"
)

type Transaction struct {
	ID      []byte
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

type TransactionInput struct {
	ID        []byte
	Output    int
	Signiture string
}

type TransactionOutput struct {
	Value     int
	PublicKey string
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [64]byte

	encoder := gob.NewEncoder(&encoded)
	if err := encoder.Encode(tx); err != nil {
		panic(err)
	}
	hash = sha512.Sum512(encoded.Bytes())
	tx.ID = hash[:]
}

func CoinbaseTransaction(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("coins to %s", to)
	}

	txin := TransactionInput{[]byte{}, -1, data}
	txout := TransactionOutput{100, to}

	transaction := Transaction{nil, []TransactionInput{txin}, []TransactionOutput{txout}}
	transaction.SetID()

	return &transaction
}

func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Output == -1
}

func (in *TransactionInput) CanUnlock(data string) bool {
	return in.Signiture == data
}

func (out *TransactionOutput) CanBeUnlocked(data string) bool {
	return out.PublicKey == data
}

func NewTransaction(from, to string, amount int, chain *Chain) *Transaction {
	var inputs []TransactionInput
	var outputs []TransactionOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)
	if acc < amount {
		panic("Fund error")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			panic(err)
		}
		for _, out := range outs {
			input := TransactionInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, TransactionOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TransactionOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
