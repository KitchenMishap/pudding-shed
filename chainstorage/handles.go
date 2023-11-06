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

func (hh HashHeight) Height() int64 {
	return hh.height
}
func (hh HashHeight) Hash() (indexedhashes.Sha256, error) {
	return hh.hash, nil
}
func (hh HashHeight) HeightSpecified() bool {
	return hh.heightSpecified
}
func (hh HashHeight) HashSpecified() bool {
	return hh.hashSpecified
}

// BlockHandle in package chainstorage implements IBlockHandle
type BlockHandle struct {
	HashHeight
}

// Check that implements
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil)

// Functions to implement IBlockHandle

func (bh BlockHandle) IsBlockHandle() {
}
func (bh BlockHandle) IsInvalid() bool {
	return bh.Height() == -1
}

// TransHandle in package chainstorage implements ITransHandle
type TransHandle struct {
	HashHeight
}

// Check that implements
var _ chainreadinterface.ITransHandle = (*TransHandle)(nil)

// Functions to implement ITransHandle

func (th TransHandle) IsTransHandle() {
}
func (th TransHandle) IsInvalid() bool {
	return th.Height() == -1
}

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

func (txi TxiHandle) ParentTrans() chainreadinterface.ITransHandle {
	return txi.TransHandle
}
func (txi TxiHandle) ParentIndex() int64 {
	return txi.index
}
func (txi TxiHandle) TxiHeight() int64 {
	return txi.txxHeight
}
func (txi TxiHandle) ParentSpecified() bool {
	return txi.transIndexSpecified
}
func (txi TxiHandle) TxiHeightSpecified() bool {
	return txi.txxHeightSpecified
}

// TxoHandle implements ITxoHandle
type TxoHandle struct {
	TxxHandle
}

// Check that implements
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil)

// Functions to implement ITxoHandle

func (txo TxoHandle) ParentTrans() chainreadinterface.ITransHandle {
	return txo.TransHandle
}
func (txo TxoHandle) ParentIndex() int64 {
	return txo.index
}
func (txo TxoHandle) TxoHeight() int64 {
	return txo.txxHeight
}
func (txo TxoHandle) ParentSpecified() bool {
	return txo.transIndexSpecified
}
func (txo TxoHandle) TxoHeightSpecified() bool {
	return txo.txxHeightSpecified
}

func InvalidBlock() BlockHandle {
	return BlockHandle{HashHeight{height: -1, hashSpecified: false, heightSpecified: true}}
}
func InvalidTrans() TransHandle {
	return TransHandle{HashHeight{height: -1, hashSpecified: false, heightSpecified: true}}
}
