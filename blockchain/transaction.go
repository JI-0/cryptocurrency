package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/JI-0/private-cryptocurrency/wallet"
	"github.com/cloudflare/circl/sign/ed448"
)

type Transaction struct {
	ID      []byte
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

type TransactionOutput struct {
	Value         int
	PublicKeyHash []byte
}

type TransactionOutputs struct {
	Outputs []TransactionOutput
}

type TransactionInput struct {
	ID        []byte
	Output    int
	Signature []byte
	PublicKey []byte
}

func CoinbaseTransaction(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 512)
		if _, err := rand.Read(randData); err != nil {
			panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	hash := sha512.Sum512([]byte(data))
	priv, err := os.ReadFile("keys/master_ed448.priv")
	if err != nil {
		panic(err)
	}
	signiture := ed448.Sign(priv, hash[:], "")
	txin := TransactionInput{hash[:], -1, signiture, nil}
	txout := NewTxOutput(100, to)

	transaction := Transaction{nil, []TransactionInput{txin}, []TransactionOutput{*txout}}
	transaction.ID = transaction.Hash()

	return &transaction
}

func (tx *Transaction) IsCoinbaseTransaction() bool {
	if len(tx.Inputs) == 1 && tx.Inputs[0].Output < 0 {
		pub, err := os.ReadFile("keys/master_ed448.pub")
		if err != nil {
			panic(err)
		}
		return ed448.Verify(pub, tx.Inputs[0].ID, tx.Inputs[0].Signature, "")
	}
	return false
}

func NewTransaction(w *wallet.Wallet, to string, amount int, UTXOs *UTXOSet) *Transaction {
	var inputs []TransactionInput
	var outputs []TransactionOutput

	publicKeyHash := wallet.PublicKeyHash(w.PublicKey)

	acc, validOutputs := UTXOs.FindSpendableOutputs(publicKeyHash, amount)
	if acc < amount {
		panic("Fund error")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			panic(err)
		}
		for _, out := range outs {
			input := TransactionInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, *NewTxOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, string(w.Address())))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXOs.Chain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func (tx *Transaction) Hash() []byte {
	var hash [64]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha512.Sum512(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TransactionInput
	var outputs []TransactionOutput
	for _, in := range tx.Inputs {
		inputs = append(inputs, TransactionInput{in.ID, in.Output, nil, nil})
	}
	for _, out := range tx.Outputs {
		outputs = append(outputs, TransactionOutput{out.Value, out.PublicKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, previousTxs map[string]Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}

	for _, in := range tx.Inputs {
		if previousTxs[hex.EncodeToString(in.ID)].ID == nil {
			panic("previous transaction input does not exist")
		}
	}
	txCopy := tx.TrimmedCopy()
	for inId, in := range txCopy.Inputs {
		previousTx := previousTxs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PublicKey = previousTx.Outputs[in.Output].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		if err != nil {
			panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature
	}
}

func (tx *Transaction) Verify(previousTxs map[string]Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}

	for _, in := range tx.Inputs {
		if previousTxs[hex.EncodeToString(in.ID)].ID == nil {
			panic("previous transaction input does not exist")
		}
	}
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P521()
	for inId, in := range tx.Inputs {
		previousTx := previousTxs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PublicKey = previousTx.Outputs[in.Output].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PublicKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])
		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PublicKey)
		x.SetBytes(in.PublicKey[:(keyLen / 2)])
		y.SetBytes(in.PublicKey[(keyLen / 2):])

		rawPublicKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPublicKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}
	return true
}

func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("		Input %d:", i))
		lines = append(lines, fmt.Sprintf("			TXID: %x", input.ID))
		lines = append(lines, fmt.Sprintf("			Out: %d", input.Output))
		lines = append(lines, fmt.Sprintf("			Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("			PublicKey: %x", input.PublicKey))
	}
	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("		Output %d:", i))
		lines = append(lines, fmt.Sprintf("			Value: %d", output.Value))
		lines = append(lines, fmt.Sprintf("			PublicKeyHash: %x", output.PublicKeyHash))
	}
	return strings.Join(lines, "\n")
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	encoder := gob.NewEncoder(&encoded)
	if err := encoder.Encode(tx); err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&transaction); err != nil {
		panic(err)
	}
	return transaction
}

// Transaction input
func (in *TransactionInput) UsesKey(publicKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PublicKey)
	return bytes.Compare(lockingHash, publicKeyHash) == 0
}

// Transaction output
func NewTxOutput(value int, address string) *TransactionOutput {
	txo := &TransactionOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

func (out *TransactionOutput) Lock(address []byte) {
	publicKeyHash := wallet.Base58Decode(address)
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-wallet.ChecksumLen]
	out.PublicKeyHash = publicKeyHash
}

func (out *TransactionOutput) IsLockedWithKey(publicKeyHash []byte) bool {
	return bytes.Compare(out.PublicKeyHash, publicKeyHash) == 0
}

func (outs *TransactionOutputs) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(outs); err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func DeserializeOutputs(data []byte) TransactionOutputs {
	var outputs TransactionOutputs
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&outputs); err != nil {
		panic(err)
	}
	return outputs
}
