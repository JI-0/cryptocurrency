package blockchain

import (
	"crypto/sha512"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}
	if left == nil && right == nil {
		hash := sha512.Sum512(data)
		node.Data = hash[:]
	} else {
		prevHash := append(left.Data, right.Data...)
		hash := sha512.Sum512(prevHash)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right
	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, d := range data {
		node := NewMerkleNode(nil, nil, d)
		nodes = append(nodes, *node)
	}

	for len(nodes) > 1 {
		var level []MerkleNode
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			level = append(level, *node)
		}
		nodes = level
	}
	tree := MerkleTree{&nodes[0]}
	return &tree
}
