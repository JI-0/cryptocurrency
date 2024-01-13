package wallet

import (
	"crypto/x509"
	"fmt"
	"os"
	"strings"
)

const walletsFile = "./tmp/wallets_%s"

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallets(path string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.Load(path)
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

func (ws *Wallets) Save(path string) {
	filePath := fmt.Sprintf(walletsFile, path)
	for address, wallet := range ws.Wallets {
		privateKeyBuffer, err := x509.MarshalECPrivateKey(&wallet.PrivateKey)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(filePath+address+".priv", privateKeyBuffer, 0644); err != nil {
			panic(err)
		}
		if err := os.WriteFile(filePath+address+".pub", wallet.PublicKey, 0644); err != nil {
			panic(err)
		}
	}
}

func (ws *Wallets) Load(path string) error {
	filePath := fmt.Sprintf(walletsFile, path)
	files, err := os.ReadDir(filePath)
	if err != nil {
		return err
	}
	for _, file := range files {
		name := file.Name()
		if strings.Contains(name, ".priv") {
			address := name[:strings.IndexByte(name, '.')]
			println("HERE", address)
			privateKeyBuffer, err := os.ReadFile(filePath + address + ".priv")
			if err != nil {
				println("CRASH1", err.Error())
				continue
			}
			privateKey, err := x509.ParseECPrivateKey(privateKeyBuffer)
			if err != nil {
				println("CRASH2", err.Error())
				continue
			}
			publicKey, err := os.ReadFile(filePath + address + ".pub")
			if err != nil {
				println("CRASH3", err.Error())
				continue
			}
			ws.Wallets[address] = OpenWallet(*privateKey, publicKey)
		}
	}
	return nil
}
