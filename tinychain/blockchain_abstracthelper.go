package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"testing"
)

func TestGenesisBlock_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	handle := blockchain.GenesisBlock()
	if handle.IsInvalid() {
		t.Error("genesis block handle cannot be invalid")
	}

	block := blockchain.BlockInterface(handle)
	if block == nil {
		t.Error("genesis block cannot be nil")
	}
	if block.IsInvalid() {
		t.Error("genesis block cannot be invalid")
	}
}

func TestGenesisParentInvalid_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	handle := blockchain.GenesisBlock()
	block := blockchain.BlockInterface(handle)
	parent := blockchain.ParentBlock(block)
	if !parent.IsInvalid() {
		t.Error("parent of genesis block must be invalid")
	}
}

func TestGenesisNextParent_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	handle := blockchain.GenesisBlock()
	nextHandle := blockchain.NextBlock(handle)
	if nextHandle.IsInvalid() {
		t.Error("next after genesis block cannot be invalid")
	}
	prev := blockchain.ParentBlock(nextHandle)
	if prev.IsInvalid() {
		t.Error("parent of next of genesis cannot be invalid")
	}
	if prev.Height() != handle.Height() {
		t.Error("parent of next block after genesis must be genesis")
	}
}

func TestGenesisTransaction_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	genesisTrans := blockchain.GenesisTransaction()
	if genesisTrans.IsInvalid() {
		t.Error("genesis transaction cannot be invalid")
	}
	prevTransaction := blockchain.PreviousTransaction(genesisTrans)
	if !prevTransaction.IsInvalid() {
		t.Error("previous of genesis transaction must be invalid")
	}
}

func TestLatestNextBlock_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	block := blockchain.LatestBlock()
	if block == nil {
		t.Error("latest block cannot be nil")
	}
	if block.IsInvalid() {
		t.Error("latest block cannot be invalid")
	}
	next := blockchain.NextBlock(block)
	if !next.IsInvalid() {
		t.Error("next after latest block should be invalid")
	}
}

func TestLatestBlockNotGenesis_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	genesisBlock := blockchain.GenesisBlock()
	latestBlock := blockchain.LatestBlock()
	if latestBlock.Height() == genesisBlock.Height() {
		t.Error("latest block should not be genesis block")
	}
}

func TestLatestPrevNextBlock_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	latestBlock := blockchain.LatestBlock()
	prevBlock := blockchain.ParentBlock(latestBlock)
	if prevBlock.Height() == latestBlock.Height() {
		t.Error("prev before latest block cannot be latest block")
	}
	nextBlock := blockchain.NextBlock(prevBlock)
	if nextBlock.Height() != latestBlock.Height() {
		t.Error("next after prev of latest block should be latest block")
	}
}

func TestLatestNextTransaction_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	transaction := blockchain.LatestTransaction()
	if transaction.IsInvalid() {
		t.Error("latest transaction cannot be invalid")
	}
	nextTransaction := blockchain.NextTransaction(transaction)
	if !nextTransaction.IsInvalid() {
		t.Error("next after latest transaction must be invalid")
	}
}
