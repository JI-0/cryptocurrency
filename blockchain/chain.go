package blockchain

type Chain struct {
	Blocks []*Block
}

func Genesis() *Block {
	return NewBlock("genesis", []byte{})
}

func NewChain() *Chain {
	return &Chain{[]*Block{Genesis()}}
}

func (c *Chain) AddBlock(data string) {
	prevBlock := c.Blocks[len(c.Blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	c.Blocks = append(c.Blocks, newBlock)
}
