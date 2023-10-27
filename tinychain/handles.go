package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// A HashHeight for tinychain that doesn't store a hash
type HashHeight struct {
	height int64
}

func (hh HashHeight) Height() int64 {
	return hh.height
}
func (hh HashHeight) Hash() indexedhashes.Sha256 {
	return indexedhashes.Sha256{}
}
func (hh HashHeight) HeightSpecified() bool {
	return true
}
func (hh HashHeight) HashSpecified() bool {
	return false
}

// BlockHandle implements IBlockHandle
type BlockHandle struct {
	HashHeight
}

func (bh BlockHandle) IsBlockHandle() {
}
func (bh BlockHandle) IsInvalid() bool {
	return bh.Height() == -1
}

// Check that implements
var _ chainreadinterface.IBlockHandle = (*BlockHandle)(nil)

// TransHandle implements ITransHandle
type TransHandle struct {
	HashHeight
}

func (th TransHandle) IsTransHandle() {
}

func (th TransHandle) IsInvalid() bool {
	return th.Height() == -1
}

// Check that implements
var _ chainreadinterface.ITransHandle = (*TransHandle)(nil)

// A TransIndex for tinychain
type TransIndex struct {
	TransHandle
	index int64
}

// A TxxHandle for tinychain
type TxxHandle struct {
	TransIndex
	txxHeight          int64
	txxHeightSpecified bool
}

// TxiHandle implements ITxiHandle
type TxiHandle struct {
	TxxHandle
}

func (th TxiHandle) ParentTrans() chainreadinterface.ITransHandle {
	return th
}
func (th TxiHandle) ParentIndex() int64 {
	return th.index
}
func (th TxiHandle) TxiHeight() int64 {
	return th.txxHeight
}
func (th TxiHandle) ParentSpecified() bool {
	return true
}
func (th TxiHandle) TxiHeightSpecified() bool {
	return true
}

// Check that implements
var _ chainreadinterface.ITxiHandle = (*TxiHandle)(nil)

// TxoHandle implements ITxoHandle
type TxoHandle struct {
	TxxHandle
}

func (th TxoHandle) ParentTrans() chainreadinterface.ITransHandle {
	return th
}
func (th TxoHandle) ParentIndex() int64 {
	return th.index
}
func (th TxoHandle) TxoHeight() int64 {
	return th.txxHeight
}
func (th TxoHandle) ParentSpecified() bool {
	return true
}
func (th TxoHandle) TxoHeightSpecified() bool {
	return true
}

func InvalidBlock() BlockHandle {
	result := BlockHandle{}
	result.height = -1
	return result
}

// Check that implements
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil)
