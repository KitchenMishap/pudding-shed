package tinychain

import (
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	block := TheTinyChain.GenesisBlock()
	if block == nil {
		t.Error()
	}
	if block.IsInvalid() {
		t.Error()
	}
}

func TestGenesisParentInvalid(t *testing.T) {
	block := TheTinyChain.GenesisBlock()
	if !TheTinyChain.ParentBlock(block).IsInvalid() {
		t.Error()
	}
}

func TestGenesisNextParent(t *testing.T) {
	block := TheTinyChain.GenesisBlock()
	next := TheTinyChain.NextBlock(block)
	if next.IsInvalid() {
		t.Error()
	}
	prev := TheTinyChain.ParentBlock(next)
	if prev.IsInvalid() {
		t.Error()
	}
	if prev != block {
		t.Error()
	}
}

func TestGenesisTransaction(t *testing.T) {
	genesisTransaction := TheTinyChain.GenesisTransaction()
	if genesisTransaction.IsInvalid() {
		t.Error()
	}
	prevTransaction := TheTinyChain.PreviousTransaction(genesisTransaction)
	if !prevTransaction.IsInvalid() {
		t.Error()
	}
}

func TestLatestNextBlock(t *testing.T) {
	block := TheTinyChain.LatestBlock()
	if block == nil {
		t.Error()
	}
	if block.IsInvalid() {
		t.Error()
	}
	next := TheTinyChain.NextBlock(block)
	if !next.IsInvalid() {
		t.Error()
	}
}

func TestLatestBlockNotGenesis(t *testing.T) {
	genesisBlock := TheTinyChain.GenesisBlock()
	latestBlock := TheTinyChain.LatestBlock()
	if latestBlock == genesisBlock {
		t.Error()
	}
}

func TestLatestPrevNextBlock(t *testing.T) {
	latestBlock := TheTinyChain.LatestBlock()
	prevBlock := TheTinyChain.ParentBlock(latestBlock)
	if prevBlock == latestBlock {
		t.Error()
	}
	nextBlock := TheTinyChain.NextBlock(prevBlock)
	if nextBlock != latestBlock {
		t.Error()
	}
}

func TestLatestNextTransaction(t *testing.T) {
	transaction := TheTinyChain.LatestTransaction()
	if transaction.IsInvalid() {
		t.Error()
	}
	nextTransaction := TheTinyChain.NextTransaction(transaction)
	if !nextTransaction.IsInvalid() {
		t.Error()
	}
}
