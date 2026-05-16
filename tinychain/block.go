package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type Block struct {
	BlockHandle
	transactions []Transaction
}

func (b *Block) TransactionCount() (int64, error) {
	return int64(len(b.transactions)), nil
}

func (b *Block) NthTransaction(n int64) (chainreadinterface.ITransHandle, error) {
	return &b.transactions[n], nil
}

func (b *Block) NonEssentialInts() (*map[string]int64, error) {
	nonEssentialInts := make(map[string]int64)
	nonEssentialInts["size"] = 285 // Make it same as real chain
	nonEssentialInts["time"] = 12345
	return &nonEssentialInts, nil
}

// Compiler check that implements
var _ chainreadinterface.IBlock = (*Block)(nil)
