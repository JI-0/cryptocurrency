package test

import (
	"fmt"
	"testing"

	"github.com/JI-0/private-cryptocurrency/wallet"
	"golang.org/x/exp/slices"
)

const (
	version        = byte(0x00)
	numberOfWalets = 100000
)

// Warning: test takes some time to finish
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

	addresses := []string{string(address)}

	for i := 0; i < numberOfWalets-1; i++ {
		w := wallet.NewWallet()
		publicHash := wallet.PublicKeyHash(w.PublicKey)
		versionedHash := append([]byte{version}, publicHash...)
		checksum := wallet.CheckSum(versionedHash)

		fullHash := append(versionedHash, checksum...)
		address := wallet.Base58Encode(fullHash)

		fmt.Printf("Address: %x\n", address)

		if slices.Contains(addresses, string(address)) {
			t.Fatalf(`Dupplicate wallet addresses`)
		}

		addresses = append(addresses, string(address))
	}
}
