package network

import "github.com/JI-0/private-cryptocurrency/blockchain"

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

var (
	nodeAddress     string
	minerAddress    string
	KnownNodes      = []string{"localhost:3000"}
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]blockchain.Transaction)
)

type Address struct {
	AddressList []string
}

type Block struct {
	AddressFrom string
	Block       []byte
}

type GetBlocks struct {
	AddressFrom string
}

type GetData struct {
	AddressFrom string
}
