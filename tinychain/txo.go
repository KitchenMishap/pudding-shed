package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// A transaction output supporting the ITxo interface
type txo struct {
	satoshis int64
}

func (atxo *txo) Satoshis() int64 {
	return atxo.satoshis
}

// Compiler check that it implements
var _ chainreadinterface.ITxo = (*txo)(nil)
