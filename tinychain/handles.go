package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// A HashHeight for tinychain that doesn't store a hash
// The hash is merely the hash of the height number
type HashHeight struct {
	height int64
}

func (hh HashHeight) Height() int64 {
	return hh.height
}
func (hh HashHeight) Hash() (indexedhashes.Sha256, error) {
	return HashOfInt(uint64(hh.height)), nil
}
func (hh HashHeight) HeightSpecified() bool {
	return true
}
func (hh HashHeight) HashSpecified() bool {
	return true
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

func (txi TxiHandle) ParentTrans() chainreadinterface.ITransHandle {
	return txi.TransHandle
}
func (txi TxiHandle) ParentIndex() int64 {
	return txi.index
}
func (txi TxiHandle) TxiHeight() int64 {
	return -1
}
func (txi TxiHandle) ParentSpecified() bool {
	return true
}
func (th TxiHandle) TxiHeightSpecified() bool {
	return false
}

// Check that implements
var _ chainreadinterface.ITxiHandle = (*TxiHandle)(nil)

// TxoHandle implements ITxoHandle
type TxoHandle struct {
	TxxHandle
}

func (txo TxoHandle) ParentTrans() chainreadinterface.ITransHandle {
	return txo.TransHandle
}
func (txo TxoHandle) ParentIndex() int64 {
	return txo.index
}
func (txo TxoHandle) TxoHeight() int64 {
	return -1
}
func (th TxoHandle) ParentSpecified() bool {
	return true
}
func (th TxoHandle) TxoHeightSpecified() bool {
	return false
}

func InvalidBlock() BlockHandle {
	result := BlockHandle{}
	result.height = -1
	return result
}

// Check that implements
var _ chainreadinterface.ITxoHandle = (*TxoHandle)(nil)
