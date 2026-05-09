package intrinsicobjectscri

import (
	"errors"

	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

type TxoHandle struct {
	txId   indexedhashes.Sha256
	vIndex int64
}

// intrinsicobjectscri.TxoHandle implements chainreadinterface.ITxoHandle
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil) // Check that implements

func (htxo *TxoHandle) ParentTrans() chainreadinterface.ITransHandle { return nil }
func (htxo *TxoHandle) ParentIndex() int64                           { return -1 }
func (htxo *TxoHandle) ParentSpecified() bool                        { return false }
func (htxo *TxoHandle) TxoHeight() int64                             { return -1 }
func (htxo *TxoHandle) IndicesPath() (int64, int64, int64)           { return -1, -1, -1 }
func (htxo *TxoHandle) TxoHeightSpecified() bool                     { return false }
func (htxo *TxoHandle) IndicesPathSpecified() bool                   { return false }

type AddressHandle struct {
	puddingHash3 indexedhashes.Sha256
}

// intrinsicobjectscri.AddressHandle implements chainreadinterface.IAddressHandle
var _ chainreadinterface.IAddressHandle = (*AddressHandle)(nil) // Check that implements

// Functions in jsonblock.AddressHandle to implement chainreadinterface.IAddressHandle
func (ah *AddressHandle) Hash() indexedhashes.Sha256 { return ah.puddingHash3 }
func (ah *AddressHandle) Height() int64              { return -1 }
func (ah *AddressHandle) HashSpecified() bool        { return true }
func (ah *AddressHandle) HeightSpecified() bool      { return false }

// Sneakily, intrinsicchain.AddressHandle also implements chainreadinterface.IAddress, with limited functionality
var _ chainreadinterface.IAddress = (*AddressHandle)(nil) // Check that implements
func (ah *AddressHandle) TxoCount() (int64, error) {
	return -1, errors.New("intrinsicchain.AddressHandle.TxoCount(): not supported")
}
func (ah *AddressHandle) NthTxo(_ int64) (chainreadinterface.ITxoHandle, error) {
	return nil, errors.New("intrinsicchain.AddressHandle.NthTxo(): not supported")
}
