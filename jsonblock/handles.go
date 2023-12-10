package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// BlockHandle for package jsonblock implements IBlockHandle
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil) // Check that implements
type BlockHandle struct {
	hash   indexedhashes.Sha256
	height int64
}

func (bh *BlockHandle) Height() int64                       { return bh.height }
func (bh *BlockHandle) Hash() (indexedhashes.Sha256, error) { return bh.hash, nil }
func (bh *BlockHandle) HeightSpecified() bool               { return true }
func (bh *BlockHandle) HashSpecified() bool                 { return true }
func (bh *BlockHandle) IsBlockHandle()                      {}
func (bh *BlockHandle) IsInvalid() bool                     { return bh.Height() == -1 }

// jsonblock.TransHandle implements chainreadinterface.ITransHandle
var _ chainreadinterface.ITransHandle = (*TransHandle)(nil) // Check that implements
type TransHandle struct {
	blockHeight int64
	nthInBlock  int64 // -1 indicates invalid handle
}

func (th *TransHandle) Height() int64                       { return -1 } // Remember Height() is a TRANSACTION height
func (th *TransHandle) Hash() (indexedhashes.Sha256, error) { return indexedhashes.Sha256{}, nil }
func (th *TransHandle) IndicesPath() (int64, int64)         { return th.blockHeight, th.nthInBlock }
func (th *TransHandle) HeightSpecified() bool               { return false }
func (th *TransHandle) HashSpecified() bool                 { return false }
func (th *TransHandle) IndicesPathSpecified() bool          { return true }
func (th *TransHandle) IsTransHandle()                      {}
func (th *TransHandle) IsInvalid() bool                     { return th.nthInBlock == -1 }

// TxxIndicesPath specifies block height, nth transaction within block, and vin/vout index (vindex)
// It is used for TxiHandle and TxoHandle
type TxxIndicesPath struct {
	transHandle TransHandle
	vIndex      int64
}

// jsonblock.TxiHandle implements chainreadinterface.ITxiHandle
var _ chainreadinterface.ITxiHandle = (*TxiHandle)(nil) // Check that implements
type TxiHandle struct {
	TxxIndicesPath
}

// Functions in jsonblock.TxiHandle to implement chainreadinterface.ITxiHandle

func (th *TxiHandle) ParentTrans() chainreadinterface.ITransHandle { return &th.transHandle }
func (th *TxiHandle) ParentIndex() int64                           { return th.vIndex }
func (th *TxiHandle) TxiHeight() int64                             { return -1 }
func (th *TxiHandle) IndicesPath() (int64, int64, int64) {
	return th.transHandle.blockHeight, th.transHandle.nthInBlock, th.vIndex
}
func (th *TxiHandle) ParentSpecified() bool      { return true }
func (th *TxiHandle) TxiHeightSpecified() bool   { return false }
func (th *TxiHandle) IndicesPathSpecified() bool { return true }

// jsonblock.TxoHandle implements chainreadinterface.ITxoHandle
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil) // Check that implements
type TxoHandle struct {
	TxxIndicesPath
}

// Functions in jsonblock.TxoHandle to implement chainreadinterface.ITxoHandle

func (th *TxoHandle) ParentTrans() chainreadinterface.ITransHandle { return &th.transHandle }
func (th *TxoHandle) ParentIndex() int64                           { return th.vIndex }
func (th *TxoHandle) TxoHeight() int64                             { return -1 }
func (th *TxoHandle) IndicesPath() (int64, int64, int64) {
	return th.transHandle.blockHeight, th.transHandle.nthInBlock, th.vIndex
}
func (th *TxoHandle) ParentSpecified() bool      { return true }
func (th *TxoHandle) TxoHeightSpecified() bool   { return false }
func (th *TxoHandle) IndicesPathSpecified() bool { return true }
