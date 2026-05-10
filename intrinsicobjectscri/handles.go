package intrinsicobjectscri

import (
	"errors"

	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// intrinsicobjectscri.BlockHandle implements chainreadinterface.IBlockHandle
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil) // Check that implements
type BlockHandle struct {
	hash      indexedhashes.Sha256 // intrinsic blocks do know their hash
	isInvalid bool
}

func (bh *BlockHandle) Height() int64                       { return -1 }
func (bh *BlockHandle) Hash() (indexedhashes.Sha256, error) { return bh.hash, nil }
func (bh *BlockHandle) HeightSpecified() bool               { return false }
func (bh *BlockHandle) HashSpecified() bool                 { return true }
func (bh *BlockHandle) IsBlockHandle()                      {}
func (bh *BlockHandle) IsInvalid() bool                     { return bh.isInvalid }

type TransHandle struct {
	transactionId indexedhashes.Sha256
	vOut          int64 // -1 indicates invalid handle
}

// intrinsicchaincri.TransHandle implements chainreadinterface.ITransHandle
var _ chainreadinterface.ITransHandle = (*TransHandle)(nil) // Check that implements

func (th *TransHandle) Height() int64                       { return -1 } // Remember Height() is a TRANSACTION height
func (th *TransHandle) Hash() (indexedhashes.Sha256, error) { return th.transactionId, nil }
func (th *TransHandle) IndicesPath() (int64, int64)         { return -1, -1 }
func (th *TransHandle) HeightSpecified() bool               { return false }
func (th *TransHandle) HashSpecified() bool                 { return true }
func (th *TransHandle) IndicesPathSpecified() bool          { return false }
func (th *TransHandle) IsTransHandle()                      {}
func (th *TransHandle) IsInvalid() bool                     { return th.vOut == -1 }

type TxiHandle struct {
	txId   indexedhashes.Sha256
	vIndex int64
}

// intrinsicobjectscri.TxiHandle implements chainreadinterface.ITxiHandle
var _ chainreadinterface.ITxiHandle = (*TxiHandle)(nil) // Check that implements

func (th *TxiHandle) ParentTrans() chainreadinterface.ITransHandle {
	result := TransHandle{}
	result.transactionId = th.txId
	return &result
}
func (th *TxiHandle) ParentIndex() int64                 { return th.vIndex }
func (th *TxiHandle) TxiHeight() int64                   { return -1 }
func (th *TxiHandle) IndicesPath() (int64, int64, int64) { return -1, -1, -1 }
func (th *TxiHandle) ParentSpecified() bool              { return true }
func (th *TxiHandle) TxiHeightSpecified() bool           { return false }
func (th *TxiHandle) IndicesPathSpecified() bool         { return false }

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
