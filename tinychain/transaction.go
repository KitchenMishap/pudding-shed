package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// Transaction implements ITransaction
type Transaction struct {
	TransHandle
	txis []Txi
	txos []Txo
}

func (t Transaction) TxiCount() (int64, error) {
	return int64(len(t.txis)), nil
}

func (t Transaction) NthTxi(n int64) (chainreadinterface.ITxiHandle, error) {
	return t.txis[n], nil
}

func (t Transaction) TxoCount() (int64, error) {
	return int64(len(t.txos)), nil
}

func (t Transaction) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) {
	return t.txos[n], nil
}

// Compiler check that implements
var _ chainreadinterface.ITransaction = (*Transaction)(nil)
