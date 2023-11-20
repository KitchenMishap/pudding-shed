package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// A HashHeight for package jsonblock that stores hash and height
type HashHeight struct {
	height int64
	hash   indexedhashes.Sha256
}

func (hh *HashHeight) Height() int64                       { return hh.height }
func (hh *HashHeight) Hash() (indexedhashes.Sha256, error) { return hh.hash, nil }
func (hh *HashHeight) HeightSpecified() bool               { return true }
func (hh *HashHeight) HashSpecified() bool                 { return true }

// BlockHandle for package jsonblock implements IBlockHandle
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil) // Check that implements
type BlockHandle struct {
	HashHeight
}

func (bh *BlockHandle) IsBlockHandle()  {}
func (bh *BlockHandle) IsInvalid() bool { return bh.Height() == -1 }

// Check that implements
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil)

// jsonblock.TransHandle implements chainreadinterface.ITransHandle
var _ chainreadinterface.ITransHandle = (*TransHandle)(nil) // Check that implements
type TransHandle struct {
	hash           indexedhashes.Sha256
	parentBlock    BlockHandle
	nthInBlock     int64 // -1 indicates invalid handle
	blockSpecified bool  // The above information won't always be specified
}

func (th *TransHandle) Height() int64                       { return -1 }
func (th *TransHandle) Hash() (indexedhashes.Sha256, error) { return th.hash, nil }
func (th *TransHandle) HeightSpecified() bool               { return false }
func (th *TransHandle) HashSpecified() bool                 { return true }
func (th *TransHandle) IsTransHandle()                      {}
func (th *TransHandle) IsInvalid() bool                     { return th.nthInBlock == -1 }

// jsonblock.Txi ALSO handles chainreadinterface.ITxiHandle, so see file transactions.go for that

// jsonblock.TxoHandle implements chainreadinterface.ITxoHandle
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil) // Check that implements
type TxoHandle struct {
	txid indexedhashes.Sha256
	vout int64
}

// functions in jsonblock.TxoHandle to implement chainreadinterface.ITxoHandle

func (txo *TxoHandle) ParentTrans() chainreadinterface.ITransHandle {
	var th TransHandle
	th.hash = txo.txid
	th.blockSpecified = false
	return &th
}
func (txo *TxoHandle) ParentIndex() int64       { return txo.vout }
func (txo *TxoHandle) TxoHeight() int64         { return -1 }
func (txo *TxoHandle) ParentSpecified() bool    { return true }
func (txo *TxoHandle) TxoHeightSpecified() bool { return false }
