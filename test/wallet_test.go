package test

import (
	"crypto/x509"
	"fmt"
	"os"
	"testing"

	"github.com/JI-0/private-cryptocurrency/wallet"
	"golang.org/x/exp/slices"
)

const (
	walletFile     = "./tmp/wallets/"
	version        = byte(0x00)
	numberOfWalets = 10000
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

func TestWalletCreationAndSavingLoading(t *testing.T) {
	// Delete save file from previous test
	if err := os.RemoveAll(walletFile); err != nil {
		t.Fatal("Wallet file error: ", err)
	}
	if err := os.Mkdir(walletFile, 0700); err != nil {
		t.Fatal("Cannot create dir")
	}
	//Create wallets and load them
	wallets, err := wallet.NewWallets()
	if err != nil {
		t.Fatal(err)
	}
	address := wallets.AddWallet()
	wallets.Save()
	println("New address: %s", address)
	wallets1, err := wallet.NewWallets()
	if err != nil {
		t.Fatal(err)
	}
	//Check public key match
	if string(wallets.Wallets[address].PublicKey) != string(wallets1.Wallets[address].PublicKey) {
		t.Fatal("Public keys do not match")
	}
	//Check private key match
	privateKeyBuffer, err := x509.MarshalECPrivateKey(&wallets.Wallets[address].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	privateKeyBuffer1, err := x509.MarshalECPrivateKey(&wallets1.Wallets[address].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	if string(privateKeyBuffer) != string(privateKeyBuffer1) {
		t.Fatal("Private keys do not match")
	}

	fmt.Printf("Wallet1 public key: %s\n", string(wallets.Wallets[address].PublicKey))
	fmt.Printf("Wallet2 public key: %s\n", string(wallets1.Wallets[address].PublicKey))
	fmt.Printf("Wallet1 private key: %s\n", string(privateKeyBuffer))
	fmt.Printf("Wallet2 private key: %s\n", string(privateKeyBuffer1))
}
