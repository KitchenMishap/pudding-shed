package jsonblock

import (
	"fmt"
)

// OneBlockHolder provides a HashesBlockchain restricted to accessing blocks in sequence.
// GenesisBlock() must be called first, and only once. At this point OneBlockHolder takes a
// JsonBlockEssential* from its InChan and presumes it to be the genesis block.
// Subsequent calls to NextBlock must be in sequence, and causes another JsonBlockEssentials* to be
// taken from InChan. Out of sequence blocks will generate errors.
type OneBlockHolderHashes struct {
	InChan                   chan *JsonBlockHashes
	currentBlock             *JsonBlockHashes
	latestBlockVisited       int64
	latestTransactionVisited int64
}

func CreateOneBlockHolderHashes() *OneBlockHolderHashes {
	res := OneBlockHolderHashes{
		InChan:                   make(chan *JsonBlockHashes),
		currentBlock:             nil,
		latestBlockVisited:       -1,
		latestTransactionVisited: -1,
	}
	return &res
}

func (obh *OneBlockHolderHashes) GenesisBlock() *JsonBlockHashes {
	if obh.latestBlockVisited != -1 {
		panic("OneBlockHolder: Can only visit Genesis block once")
	}
	fmt.Println("Attempting to receive genesis block...")
	obh.currentBlock = <-obh.InChan
	fmt.Println("...Received genesis block")

	if obh.currentBlock == nil {
		panic("OneBlockHolder: First block was nil")
	}
	if obh.currentBlock.J_height != 0 {
		panic("OneBlockHolder: First block was not height zero")
	}
	obh.latestBlockVisited = 0
	return obh.currentBlock
}

func (obh *OneBlockHolderHashes) NextBlock() (*JsonBlockHashes, error) {
	obh.currentBlock = <-obh.InChan
	obh.latestBlockVisited = int64(obh.currentBlock.J_height)
	return obh.currentBlock, nil
}
