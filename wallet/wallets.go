package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"os"
)

const walletsFile = "./tmp/wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.Load()
	return &wallets, err
}

func (ws *Wallets) AddWallet() string {
	wallet := NewWallet()
	address := string(wallet.Address())
	ws.Wallets[address] = wallet
	return address
}

func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws *Wallets) Save() {
	var content bytes.Buffer

	gob.Register(elliptic.P521())
	encoder := gob.NewEncoder(&content)
	if err := encoder.Encode(ws); err != nil {
		panic(err)
	}
	if err := os.WriteFile(walletsFile, content.Bytes(), 0644); err != nil {
		panic(err)
	}
}

func (ws *Wallets) Load() error {
	if _, err := os.Stat(walletsFile); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets
	fileContent, err := os.ReadFile(walletsFile)
	if err != nil {
		return err
	}

	gob.Register(elliptic.P521())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err := decoder.Decode(&wallets); err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets
	return nil
}
