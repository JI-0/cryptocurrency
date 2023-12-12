package test

import (
	"fmt"
	"testing"

	"github.com/JI-0/private-cryptocurrency/wallet"
)

const version = byte(0x00)

func TestWalletCreationAndAddressGeneration(t *testing.T) {
	w := wallet.NewWallet()
	publicHash := wallet.PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{version}, publicHash...)
	checksum := wallet.CheckSum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := wallet.Base58Encode(fullHash)

	fmt.Printf("Public key: %x\n", w.PublicKey)
	fmt.Printf("Public hash: %x\n", publicHash)
	fmt.Printf("Address: %x\n", address)
}
