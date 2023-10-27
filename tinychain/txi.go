package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
)

type Txi struct {
	TxiHandle
	sourceTxo TxoHandle
}

func (atxi Txi) SourceTxo() chainreadinterface.ITxoHandle {
	return atxi.sourceTxo
}

// Compiler check that it implements
var _ chainreadinterface.ITxi = (*Txi)(nil)
