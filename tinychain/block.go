package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type Block struct {
	BlockHandle
	transactions []Transaction
}

func (b Block) TransactionCount() int64 {
	return int64(len(b.transactions))
}

func (b Block) NthTransaction(n int64) chainreadinterface.ITransHandle {
	return b.transactions[n]
}

// Compiler check that implements
var _ chainreadinterface.IBlock = (*Block)(nil)
