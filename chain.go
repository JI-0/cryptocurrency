package main

type Chain struct {
	blocks []*Block
}

func Genesis() *Block {
	return NewBlock("genesis", []byte{})
}

func NewChain() *Chain {
	return &Chain{[]*Block{Genesis()}}
}

func (c *Chain) addBlock(data string) {
	prevBlock := c.blocks[len(c.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	c.blocks = append(c.blocks, newBlock)
}
