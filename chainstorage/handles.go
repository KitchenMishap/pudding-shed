package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// HashHeight in package chainstorage
type HashHeight struct {
	height          int64
	hash            indexedhashes.Sha256
	heightSpecified bool
	hashSpecified   bool
}

// BlockHandle in package chainstorage implements IBlockHandle
type BlockHandle struct {
	HashHeight
	data *concreteReadableChain
}

// Check that implements
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil)

// Functions to implement IBlockHandle

func (bh *BlockHandle) Height() int64 {
	if bh.heightSpecified {
		return bh.height
	} else {
		// ToDo Height() needs to return error
		height, _ := bh.data.blkHashes.IndexOfHash(&bh.hash)
		return height
	}
}
func (bh *BlockHandle) Hash() (indexedhashes.Sha256, error) {
	if bh.hashSpecified {
		return bh.hash, nil
	} else {
		hash := indexedhashes.Sha256{}
		err := bh.data.blkHashes.GetHashAtIndex(bh.height, &hash)
		return hash, err
	}
}
func (bh *BlockHandle) HeightSpecified() bool {
	return true
}
func (bh *BlockHandle) HashSpecified() bool {
	return true
}

func (bh *BlockHandle) IsBlockHandle() {
}
func (bh *BlockHandle) IsInvalid() bool {
	return bh.Height() == -1
}

// TransHandle in package chainstorage implements ITransHandle
type TransHandle struct {
	HashHeight
	data *concreteReadableChain
}

// Check that implements
var _ chainreadinterface.ITransHandle = (*TransHandle)(nil)

// Functions to implement ITransHandle

func (th *TransHandle) Height() int64 {
	if th.heightSpecified {
		return th.height
	} else {
		// ToDo Height() needs to return error
		height, _ := th.data.trnHashes.IndexOfHash(&th.hash)
		return height
	}
}
func (th *TransHandle) Hash() (indexedhashes.Sha256, error) {
	if th.hashSpecified {
		return th.hash, nil
	} else {
		hash := indexedhashes.Sha256{}
		err := th.data.trnHashes.GetHashAtIndex(th.height, &hash)
		return hash, err
	}
}
func (th *TransHandle) HeightSpecified() bool {
	return true
}
func (th *TransHandle) HashSpecified() bool {
	return true
}
func (th *TransHandle) IndicesPath() (int64, int64) { return -1, -1 }
func (th *TransHandle) IndicesPathSpecified() bool  { return false }
func (th *TransHandle) IsTransHandle()              {}
func (th *TransHandle) IsInvalid() bool             { return th.Height() == -1 }

// TransIndex in package chainstorage
type TransIndex struct {
	TransHandle
	index int64
}

// TxxHandle in package chainstorage
type TxxHandle struct {
	TransIndex
	txxHeight           int64
	transIndexSpecified bool
	txxHeightSpecified  bool
}

// TxiHandle in package chainstorage implements ITxiHandle
type TxiHandle struct {
	TxxHandle
}

// Check that implements
var _ chainreadinterface.ITxiHandle = (*TxiHandle)(nil)

// Functions to implement ITxiHandle

func (txi *TxiHandle) ParentTrans() chainreadinterface.ITransHandle {
	return &txi.TransHandle
}
func (txi *TxiHandle) ParentIndex() int64 {
	return txi.index
}
func (txi *TxiHandle) TxiHeight() int64 {
	return txi.txxHeight
}
func (txi *TxiHandle) ParentSpecified() bool {
	return txi.transIndexSpecified
}
func (txi *TxiHandle) TxiHeightSpecified() bool {
	return txi.txxHeightSpecified
}
func (txi *TxiHandle) IndicesPath() (int64, int64, int64) { return -1, -1, -1 }
func (txi *TxiHandle) IndicesPathSpecified() bool         { return false }

// TxoHandle implements ITxoHandle
type TxoHandle struct {
	TxxHandle
}

// Check that implements
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil)

// Functions to implement ITxoHandle

func (txo *TxoHandle) ParentTrans() chainreadinterface.ITransHandle {
	return &txo.TransHandle
}
func (txo *TxoHandle) ParentIndex() int64 {
	return txo.index
}
func (txo *TxoHandle) TxoHeight() int64 {
	return txo.txxHeight
}
func (txo *TxoHandle) ParentSpecified() bool {
	return txo.transIndexSpecified
}
func (txo *TxoHandle) TxoHeightSpecified() bool {
	return txo.txxHeightSpecified
}
func (txo *TxoHandle) IndicesPath() (int64, int64, int64) { return -1, -1, -1 }
func (txo *TxoHandle) IndicesPathSpecified() bool         { return false }
