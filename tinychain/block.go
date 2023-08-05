package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type block struct {
	height       int64
	transactions []transaction
}

func (b *block) BlockHandle() chainreadinterface.HBlock {
	return theHandles.HBlockFromHeight(b.height)
}

func (b *block) BlockHeight() int64 {
	return b.height
}

func (b *block) TransactionCount() int64 {
	return int64(len(b.transactions))
}

func (b *block) NthTransactionHandle(n int64) chainreadinterface.HTransaction {
	return theHandles.HTransactionFromHeight(b.transactions[n].height)
}

// Compiler check that implements
var _ chainreadinterface.IBlock = (*block)(nil)
