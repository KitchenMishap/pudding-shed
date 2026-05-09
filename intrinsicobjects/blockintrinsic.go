package intrinsicobjects

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Intrinsic objects may not hold data inferred from external representations.
// For example, a block may refer to the block hash of the previous block, because that info is available in the block binary.
// But a block may not hold its blockheight, because (in general, pre BIP34) that is not mentioned in the block binary.

type Block struct {
	BlockHash    indexedhashes.Sha256
	Version      uint32
	PrevHash     indexedhashes.Sha256
	MerkleRoot   indexedhashes.Sha256
	Time         uint32 // mediantime must be inferred from a block history
	NBits        uint32
	Nonce        uint32
	Transactions []Transaction

	NonEssentialIntsMap map[string]int64 // ToDo populate this
}

// intrinsicobjects.Block implements chainreadinterface.IBlockHandle
var _ chainreadinterface.IBlockHandle = (*Block)(nil) // Check that implements

func (b *Block) Height() int64                       { return -1 }
func (b *Block) HeightSpecified() bool               { return false }
func (b *Block) Hash() (indexedhashes.Sha256, error) { return b.BlockHash, nil }
func (b *Block) HashSpecified() bool                 { return true }
func (b *Block) IsBlockHandle()                      {}
func (b *Block) IsInvalid() bool                     { return false }

// intrinsicobjects.block implements chainreadinterface.IBlock
var _ chainreadinterface.IBlock = (*Block)(nil) // Check that implements

func (b *Block) TransactionCount() (int64, error) { return int64(len(b.Transactions)), nil }
func (b *Block) NthTransaction(n int64) (chainreadinterface.ITransHandle, error) {
	return &b.Transactions[n], nil
}
func (b *Block) NonEssentialInts() (*map[string]int64, error) { return &b.NonEssentialIntsMap, nil }
