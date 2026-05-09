package intrinsicobjectscri

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
)

// A Block object with an intrinsicobjects.Block, with adornments to
// implement chainreadinterface Interfaces

type Block struct {
	intrinsic    *intrinsicobjects.Block
	transactions []Transaction
	neisMap      map[string]int64 // Non Essential Ints
}

func NewBlock(intrinsic *intrinsicobjects.Block) *Block {
	result := Block{}
	result.intrinsic = intrinsic

	result.transactions = make([]Transaction, len(intrinsic.Transactions))
	for i := range intrinsic.Transactions {
		result.transactions[i] = *NewTransaction(&intrinsic.Transactions[i])
	}

	result.neisMap = make(map[string]int64)
	// ToDo

	return &result
}

// intrinsicobjectscri.Block implements chainreadinterface.IBlock
var _ chainreadinterface.IBlock = (*Block)(nil) // Check that implements

func (b *Block) TransactionCount() (int64, error) { return int64(len(b.transactions)), nil }
func (b *Block) NthTransaction(n int64) (chainreadinterface.ITransHandle, error) {
	return &b.transactions[n], nil
}
func (b *Block) NonEssentialInts() (*map[string]int64, error) { return &b.neisMap, nil }

// intrinsicobjectscri.Block also implements chainreadinterface.IBlockHandle
var _ chainreadinterface.IBlockHandle = (*Block)(nil) // Check that implements

func (b *Block) Height() int64                       { return -1 }
func (b *Block) HeightSpecified() bool               { return false }
func (b *Block) Hash() (indexedhashes.Sha256, error) { return b.intrinsic.BlockHash, nil }
func (b *Block) HashSpecified() bool                 { return true }
func (b *Block) IsBlockHandle()                      {}
func (b *Block) IsInvalid() bool                     { return false }
