package intrinsicobjects

import (
	"errors"

	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// BlockHandle for package intrinsicchain implements IBlockHandle
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil) // Check that implements
type BlockHandle struct {
	hash   indexedhashes.Sha256 // intrinsic blocks do know their hash
	height int64                // block heights can easily be inferred, so are included
}

func (bh *BlockHandle) Height() int64                       { return bh.height }
func (bh *BlockHandle) Hash() (indexedhashes.Sha256, error) { return bh.hash, nil }
func (bh *BlockHandle) HeightSpecified() bool               { return true }
func (bh *BlockHandle) HashSpecified() bool                 { return true }
func (bh *BlockHandle) IsBlockHandle()                      {}
func (bh *BlockHandle) IsInvalid() bool                     { return bh.Height() == -1 }

/*
// We give TWO implementations of chainreadinterface.ITransHandle

	type TransHandleIndices struct {
		blockHeight int64
		nthInBlock  int64 // -1 indicates invalid handle
	}
*/
type TransHandleHash struct {
	transactionId indexedhashes.Sha256
	vOut          int64 // -1 indicates invalid handle
}

/*
// intrinsicchain.TransHandleIndices implements chainreadinterface.ITransHandle
var _ chainreadinterface.ITransHandle = (*TransHandleIndices)(nil) // Check that implements

func (th *TransHandleIndices) Height() int64                       { return -1 } // Remember Height() is a TRANSACTION height
func (th *TransHandleIndices) Hash() (indexedhashes.Sha256, error) { return indexedhashes.Sha256{}, nil }
func (th *TransHandleIndices) IndicesPath() (int64, int64)         { return th.blockHeight, th.nthInBlock }
func (th *TransHandleIndices) HeightSpecified() bool               { return false }
func (th *TransHandleIndices) HashSpecified() bool                 { return false }
func (th *TransHandleIndices) IndicesPathSpecified() bool          { return true }
func (th *TransHandleIndices) IsTransHandle()                      {}
func (th *TransHandleIndices) IsInvalid() bool                     { return th.nthInBlock == -1 }
*/

// intrinsicchain.TransHandleHash implements chainreadinterface.ITransHandle
var _ chainreadinterface.ITransHandle = (*TransHandleHash)(nil) // Check that implements

func (th *TransHandleHash) Height() int64                       { return -1 } // Remember Height() is a TRANSACTION height
func (th *TransHandleHash) Hash() (indexedhashes.Sha256, error) { return th.transactionId, nil }
func (th *TransHandleHash) IndicesPath() (int64, int64)         { return -1, -1 }
func (th *TransHandleHash) HeightSpecified() bool               { return false }
func (th *TransHandleHash) HashSpecified() bool                 { return true }
func (th *TransHandleHash) IndicesPathSpecified() bool          { return false }
func (th *TransHandleHash) IsTransHandle()                      {}
func (th *TransHandleHash) IsInvalid() bool                     { return th.vOut == -1 }

// intrinsicchain.TxiHandle implements chainreadinterface.ITxiHandle
var _ chainreadinterface.ITxiHandle = (*TxiHandle)(nil) // Check that implements
type TxiHandle struct {
	txId   indexedhashes.Sha256
	vIndex int64
}

// Functions in intrinsicchain.TxiHandle to implement chainreadinterface.ITxiHandle

func (th *TxiHandle) ParentTrans() chainreadinterface.ITransHandle {
	result := TransHandleHash{}
	result.transactionId = th.txId
	return &result
}
func (th *TxiHandle) ParentIndex() int64                 { return th.vIndex }
func (th *TxiHandle) TxiHeight() int64                   { return -1 }
func (th *TxiHandle) IndicesPath() (int64, int64, int64) { return -1, -1, -1 }
func (th *TxiHandle) ParentSpecified() bool              { return true }
func (th *TxiHandle) TxiHeightSpecified() bool           { return false }
func (th *TxiHandle) IndicesPathSpecified() bool         { return false }

// intrinsicchain.TxoHandle implements chainreadinterface.ITxoHandle
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil) // Check that implements
type TxoHandle struct {
	txId   indexedhashes.Sha256
	vIndex int64
}

// Functions in intrinsicchain.TxoHandle to implement chainreadinterface.ITxoHandle
func (th *TxoHandle) ParentTrans() chainreadinterface.ITransHandle {
	result := TransHandleHash{}
	result.transactionId = th.txId
	return &result
}
func (th *TxoHandle) ParentIndex() int64                 { return th.vIndex }
func (th *TxoHandle) TxoHeight() int64                   { return -1 }
func (th *TxoHandle) IndicesPath() (int64, int64, int64) { return -1, -1, -1 }
func (th *TxoHandle) ParentSpecified() bool              { return true }
func (th *TxoHandle) TxoHeightSpecified() bool           { return false }
func (th *TxoHandle) IndicesPathSpecified() bool         { return false }

// intrinsicchain.AddressHandle implements chainreadinterface.IAddressHandle
var _ chainreadinterface.IAddressHandle = (*AddressHandle)(nil) // Check that implements
type AddressHandle struct {
	puddingHash3 indexedhashes.Sha256
}

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
func (ah *AddressHandle) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) {
	return nil, errors.New("intrinsicchain.AddressHandle.NthTxo(): not supported")
}
