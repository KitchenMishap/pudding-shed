package intrinsicobjects

import (
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
}
