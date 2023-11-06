package chainstorage

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

type Block struct {
	height int64
	data   *concreteReadableChain
}

// Functions to implement IBlockHandle as part of IBlock

func (block Block) Height() int64 {
	return block.height
}
func (block Block) Hash() (indexedhashes.Sha256, error) {
	hash := indexedhashes.Sha256{}
	err := block.data.blkHashes.GetHashAtIndex(block.height, &hash)
	return hash, err
}
func (block Block) HeightSpecified() bool {
	return true
}
func (block Block) HashSpecified() bool {
	return true
}
func (block Block) IsBlockHandle() {}
func (block Block) IsInvalid() bool {
	return block.height == -1
}

// Functions to implement IBlock

func (block Block) TransactionCount() (int64, error) {
	blocksInChain, err := block.data.blkHashes.CountHashes()
	if err != nil {
		return -1, err
	}
	transInChain, err := block.data.trnHashes.CountHashes()
	if err != nil {
		return -1, err
	}

	blockFirstTransHeight, err := block.data.blkFirstTrans.ReadWordAt(block.height)
	if err != nil {
		return -1, err
	}
	nextBlockHeight := block.height + 1
	if nextBlockHeight < blocksInChain {
		nextBlockFirstTransHeight, err := block.data.blkFirstTrans.ReadWordAt(nextBlockHeight)
		if err != nil {
			return -1, err
		}
		return nextBlockFirstTransHeight - blockFirstTransHeight, nil
	} else {
		// There might not be a next block
		return transInChain - blockFirstTransHeight, nil
	}
}
func (block Block) NthTransaction(n int64) (chainreadinterface.ITransHandle, error) {
	blockFirstTransHeight, err := block.data.blkFirstTrans.ReadWordAt(block.height)
	if err != nil {
		return InvalidTrans(), err
	}
	transHeight := blockFirstTransHeight + n
	return TransHandle{HashHeight{height: transHeight, hashSpecified: false, heightSpecified: true}}, nil
}
