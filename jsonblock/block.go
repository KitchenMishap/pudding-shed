package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// jsonblock.jsonBlockEssential implements chainreadinterface.IBlock
var _ chainreadinterface.IBlock = (*JsonBlockEssential)(nil) // Check that implements

// Functions in jsonblock.jsonBlockEssential that implement chainreadinterface.IBlockHandle as part of chainreadinterface.IBlock

func (b *JsonBlockEssential) Height() int64                       { return int64(b.J_height) }
func (b *JsonBlockEssential) Hash() (indexedhashes.Sha256, error) { return b.hash, nil }
func (b *JsonBlockEssential) HeightSpecified() bool               { return true }
func (b *JsonBlockEssential) HashSpecified() bool                 { return true }
func (b *JsonBlockEssential) IsBlockHandle()                      {}
func (b *JsonBlockEssential) IsInvalid() bool                     { return false } // A jsonBlockEssential is always valid

// Functions in jsonblock.jsonBlockEssential that implement chainreadinterface.IBlock

func (b *JsonBlockEssential) TransactionCount() (int64, error) { return int64(len(b.J_tx)), nil }
func (b *JsonBlockEssential) NthTransaction(n int64) (chainreadinterface.ITransHandle, error) {
	var th = TransHandle{}
	th.blockHeight = int64(b.J_height)
	th.nthInBlock = n
	return &th, nil
}
