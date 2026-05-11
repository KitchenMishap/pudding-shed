package intrinsicobjectscri

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
)

// A Block object with an intrinsicobjects.Block, with adornments to
// implement chainreadinterface Interfaces

type Block struct {
	intrinsic *intrinsicobjects.Block

	transactions []*Transaction
	txidMap      map[indexedhashes.Sha256]int64
	neisMap      map[string]int64 // Non Essential Ints
	blockHeight  int64
	medianTime   uint32
}

func NewBlock(intrinsic *intrinsicobjects.Block, blockHeight int64, mediantime uint32) (*Block, error) {

	result := Block{}
	result.intrinsic = intrinsic
	result.blockHeight = blockHeight
	result.medianTime = mediantime

	result.txidMap = make(map[indexedhashes.Sha256]int64)

	result.transactions = make([]*Transaction, len(intrinsic.Transactions))
	for i := range intrinsic.Transactions {
		isCoinbase := i == 0 // First transaction of block is coinbase transaction (reward plus miners' fees)
		var err error
		result.transactions[i], err = NewTransaction(&intrinsic.Transactions[i], isCoinbase,
			&result, int64(i))
		if err != nil {
			return nil, err
		}
		result.txidMap[intrinsic.Transactions[i].TxId] = int64(i)
	}

	result.neisMap = make(map[string]int64)
	// ToDo
	result.neisMap["time"] = int64(intrinsic.Time)
	result.neisMap["mediantime"] = int64(result.medianTime)
	result.neisMap["size"] = int64(intrinsic.Size)
	result.neisMap["strippedsize"] = int64(intrinsic.StrippedSize)
	result.neisMap["weight"] = int64(intrinsic.Weight)
	result.neisMap["difficulty"] = int64(intrinsic.Difficulty) // ToDo round this for JSON and Here

	return &result, nil
}

// intrinsicobjectscri.Block implements chainreadinterface.IBlock
var _ chainreadinterface.IBlock = (*Block)(nil) // Check that implements

func (b *Block) TransactionCount() (int64, error) { return int64(len(b.transactions)), nil }
func (b *Block) NthTransaction(n int64) (chainreadinterface.ITransHandle, error) {
	return b.transactions[n], nil
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
