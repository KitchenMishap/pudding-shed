package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type transaction struct {
	height int64
	txis   []txi
	txos   []txo
}

func (t *transaction) TransactionHandle() chainreadinterface.HTransaction {
	h := handle{
		height: t.height,
	}
	return h
}

func (t *transaction) TxiCount() int64 {
	return int64(len(t.txis))
}

func (t *transaction) NthTxiInterface(n int64) chainreadinterface.ITxi {
	return &t.txis[n]
}

func (t *transaction) TxoCount() int64 {
	return int64(len(t.txos))
}

func (t *transaction) NthTxoInterface(n int64) chainreadinterface.ITxo {
	return &t.txos[n]
}

// Compiler check that implements
var _ chainreadinterface.ITransaction = (*transaction)(nil)
