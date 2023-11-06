package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// A transaction output supporting the ITxo interface
type Txo struct {
	TxoHandle
	satoshis int64
}

func (atxo Txo) Satoshis() (int64, error) {
	return atxo.satoshis, nil
}

// Compiler check that it implements
var _ chainreadinterface.ITxo = (*Txo)(nil)
