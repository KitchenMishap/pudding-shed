package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
)

type txi struct {
	sourceTransactionHeight int64
	sourceIndex             int64
}

func (atxi *txi) SourceTransaction() chainreadinterface.HTransaction {
	return theHandles.hTransactionFromHeight(atxi.sourceTransactionHeight)
}

func (atxi *txi) SourceIndex() int64 {
	return atxi.sourceIndex
}

// Compiler check that it implements
var _ chainreadinterface.ITxi = (*txi)(nil)
