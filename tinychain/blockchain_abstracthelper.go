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

	block, err := blockchain.BlockInterface(handle)
	if err != nil {
		t.Error("could not get BlockInterface from blockchain")
	}
	if block == nil {
		t.Error("genesis block cannot be nil")
	}
	if block.IsInvalid() {
		t.Error("genesis block cannot be invalid")
	}
}

func TestGenesisParentInvalid_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	handle := blockchain.GenesisBlock()
	block, err := blockchain.BlockInterface(handle)
	if err != nil {
		t.Error("could not get BlockInterface from blockchain")
	}
	parent := blockchain.ParentBlock(block)
	if !parent.IsInvalid() {
		t.Error("parent of genesis block must be invalid")
	}
}

func TestGenesisNextParent_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	handle := blockchain.GenesisBlock()
	nextHandle, err := blockchain.NextBlock(handle)
	if err != nil {
		t.Error("could not get NextBlock from blockchain")
	}
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
	genesisTrans, err := blockchain.GenesisTransaction()
	if err != nil {
		t.Error("could not get GenesisTransaction from blockchain")
	}
	if genesisTrans.IsInvalid() {
		t.Error("genesis transaction cannot be invalid")
	}
	prevTransaction := blockchain.PreviousTransaction(genesisTrans)
	if !prevTransaction.IsInvalid() {
		t.Error("previous of genesis transaction must be invalid")
	}
}

func TestLatestNextBlock_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	block, err := blockchain.LatestBlock()
	if err != nil {
		t.Error("Could not get LatestBlock from blockchain")
	}
	if block == nil {
		t.Error("latest block cannot be nil")
	}
	if block.IsInvalid() {
		t.Error("latest block cannot be invalid")
	}
	next, err := blockchain.NextBlock(block)
	if err != nil {
		t.Error("could not get NextBlock from blockchain")
	}
	if !next.IsInvalid() {
		t.Error("next after latest block should be invalid")
	}
}

func TestLatestBlockNotGenesis_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	genesisBlock := blockchain.GenesisBlock()
	latestBlock, err := blockchain.LatestBlock()
	if err != nil {
		t.Error("Could not get LatestBlock from blockchain")
	}
	if latestBlock.Height() == genesisBlock.Height() {
		t.Error("latest block should not be genesis block")
	}
}

func TestLatestPrevNextBlock_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	latestBlock, err := blockchain.LatestBlock()
	if err != nil {
		t.Error("Could not get LatestBlock from blockchain")
	}
	prevBlock := blockchain.ParentBlock(latestBlock)
	if prevBlock.Height() == latestBlock.Height() {
		t.Error("prev before latest block cannot be latest block")
	}
	nextBlock, err := blockchain.NextBlock(prevBlock)
	if err != nil {
		t.Error("could not get NextBlock from blockchain")
	}
	if nextBlock.Height() != latestBlock.Height() {
		t.Error("next after prev of latest block should be latest block")
	}
}

func TestLatestNextTransaction_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	transaction, err := blockchain.LatestTransaction()
	if err != nil {
		t.Error("could not get LatestTransaction from blockchain")
	}
	if transaction.IsInvalid() {
		t.Error("latest transaction cannot be invalid")
	}
	nextTransaction, err := blockchain.NextTransaction(transaction)
	if err != nil {
		t.Error("could not get NextTransaction from blockchain")
	}
	if !nextTransaction.IsInvalid() {
		t.Error("next after latest transaction must be invalid")
	}
}
