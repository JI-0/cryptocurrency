package wallet

const walletsFile = "./tmp/wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet
}
