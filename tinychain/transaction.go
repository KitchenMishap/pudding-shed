package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// Transaction implements ITransaction
type Transaction struct {
	TransHandle
	txis []Txi
	txos []Txo
}

func (t Transaction) TxiCount() int64 {
	return int64(len(t.txis))
}

func (t Transaction) NthTxi(n int64) chainreadinterface.ITxiHandle {
	return t.txis[n]
}

func (t Transaction) TxoCount() int64 {
	return int64(len(t.txos))
}

func (t Transaction) NthTxo(n int64) chainreadinterface.ITxoHandle {
	return t.txos[n]
}

// Compiler check that implements
var _ chainreadinterface.ITransaction = (*Transaction)(nil)
