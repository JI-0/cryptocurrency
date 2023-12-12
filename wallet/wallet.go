package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLen = 4
	version     = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P521()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, public
}

func NewWallet() *Wallet {
	private, public := NewKeyPair()
	return &Wallet{private, public}
}

func PublicKeyHash(publicKey []byte) []byte {
	publicKeyHash := sha512.Sum512(publicKey)

	hasher := ripemd160.New()
	if _, err := hasher.Write(publicKeyHash[:]); err != nil {
		log.Panic(err)
	}
	publicRipeMD := hasher.Sum(nil)

	return publicRipeMD
}

func CheckSum(payload []byte) []byte {
	hash1 := sha512.Sum512(payload)
	hash2 := sha512.Sum512(hash1[:])
	return hash2[:checksumLen]
}

func (w *Wallet) Address() []byte {
	publicHash := PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{version}, publicHash...)
	checksum := CheckSum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	fmt.Printf("Public key: %x\n", w.PublicKey)
	fmt.Printf("Public hash: %x\n", publicHash)
	fmt.Printf("Address: %x\n", address)

	return address
}
