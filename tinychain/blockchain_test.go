package tinychain

import (
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	handle := TheTinyChain.GenesisBlock()
	if handle.IsInvalid() {
		t.Error("genesis block handle cannot be invalid")
	}

	block := TheTinyChain.BlockInterface(handle)
	if block == nil {
		t.Error("genesis block cannot be nil")
	}
}

func TestGenesisParentInvalid(t *testing.T) {
	handle := TheTinyChain.GenesisBlock()
	block := TheTinyChain.BlockInterface(handle)
	parent := TheTinyChain.ParentBlock(block)
	if !parent.IsInvalid() {
		t.Error("parent of genesis block must be invalid")
	}
}

func TestGenesisNextParent(t *testing.T) {
	handle := TheTinyChain.GenesisBlock()
	nextHandle := TheTinyChain.NextBlock(handle)
	if nextHandle.IsInvalid() {
		t.Error("next after genesis block cannot be invalid")
	}
	prev := TheTinyChain.ParentBlock(nextHandle)
	if prev.IsInvalid() {
		t.Error("parent of next of genesis cannot be invalid")
	}
	if prev != handle {
		t.Error("parent of next block after genesis must be genesis")
	}
}

func TestGenesisTransaction(t *testing.T) {
	genesisTrans := TheTinyChain.GenesisTransaction()
	if genesisTrans.IsInvalid() {
		t.Error("genesis transaction cannot be invalid")
	}
	prevTransaction := TheTinyChain.PreviousTransaction(genesisTrans)
	if !prevTransaction.IsInvalid() {
		t.Error("previous of genesis transaction must be invalid")
	}
}

func TestLatestNextBlock(t *testing.T) {
	block := TheTinyChain.LatestBlock()
	if block == nil {
		t.Error("latest block cannot be nil")
	}
	if block.IsInvalid() {
		t.Error("latest block cannot be invalid")
	}
	next := TheTinyChain.NextBlock(block)
	if !next.IsInvalid() {
		t.Error("next after latest block should be invalid")
	}
}

func TestLatestBlockNotGenesis(t *testing.T) {
	genesisBlock := TheTinyChain.GenesisBlock()
	latestBlock := TheTinyChain.LatestBlock()
	if latestBlock == genesisBlock {
		t.Error("latest block should not be genesis block")
	}
}

func TestLatestPrevNextBlock(t *testing.T) {
	latestBlock := TheTinyChain.LatestBlock()
	prevBlock := TheTinyChain.ParentBlock(latestBlock)
	if prevBlock == latestBlock {
		t.Error("prev before latest block cannot be latest block")
	}
	nextBlock := TheTinyChain.NextBlock(prevBlock)
	if nextBlock != latestBlock {
		t.Error("next after prev of latest block should be latest block")
	}
}

func TestLatestNextTransaction(t *testing.T) {
	transaction := TheTinyChain.LatestTransaction()
	if transaction.IsInvalid() {
		t.Error("latest transaction cannot be invalid")
	}
	nextTransaction := TheTinyChain.NextTransaction(transaction)
	if !nextTransaction.IsInvalid() {
		t.Error("next after latest transaction must be invalid")
	}
}
