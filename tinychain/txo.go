package tinychain

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
)

// A transaction output supporting the ITxo interface
type Txo struct {
	TxoHandle
	satoshis int64
}

func (txo Txo) Satoshis() (int64, error) {
	return txo.satoshis, nil
}

func (txo Txo) Address() (chainreadinterface.IAddressHandle, error) {
	return nil, errors.New("tinychain.Txo.Address(): TinyChain does not support addresses")
}

// Compiler check that it implements
var _ chainreadinterface.ITxo = (*Txo)(nil)
