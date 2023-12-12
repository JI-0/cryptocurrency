package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"os"
)

const walletsFile = "./tmp/wallets.data"

type Wallets map[string]*Wallet

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
