package tinychain

import (
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	TestGenesisBlock_helper(TheTinyChain, t)
}

func TestGenesisParentInvalid(t *testing.T) {
	TestGenesisParentInvalid_helper(TheTinyChain, t)
}

func TestGenesisNextParent(t *testing.T) {
	TestGenesisNextParent_helper(TheTinyChain, t)
}

func TestGenesisTransaction(t *testing.T) {
	TestGenesisTransaction_helper(TheTinyChain, t)
}

func TestLatestNextBlock(t *testing.T) {
	TestLatestNextBlock_helper(TheTinyChain, t)
}

func TestLatestBlockNotGenesis(t *testing.T) {
	TestLatestBlockNotGenesis_helper(TheTinyChain, t)
}

func TestLatestPrevNextBlock(t *testing.T) {
	TestLatestPrevNextBlock_helper(TheTinyChain, t)
}

func TestLatestNextTransaction(t *testing.T) {
	TestLatestNextTransaction_helper(TheTinyChain, t)
}
