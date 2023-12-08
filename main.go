package main

import "github.com/JI-0/private-cryptocurrency/blockchain"

func main() {
	chain := blockchain.NewChain()
	chain.AddBlock("data0")
	chain.AddBlock("data1")
	chain.AddBlock("data2")
}
