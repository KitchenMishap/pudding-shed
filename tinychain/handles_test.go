package tinychain

import "testing"

func TestGenesisHandle(t *testing.T) {
	if TheTinyChain.IsBlockTree() {
		t.Error("TheTinyChain can't be a BlockTree")
	}
	genesisHandle := TheTinyChain.GenesisBlock()
	genesisBlock := TheTinyChain.BlockInterface(genesisHandle)
	if !genesisBlock.HeightSpecified() {
		t.Error("genesisBlock must specify a height")
	}
	if genesisBlock.Height() != 0 {
		t.Error("genesisBlock should be height 0")
	}
}

func TestInvalidHandle(t *testing.T) {
	genesisHandle := TheTinyChain.GenesisBlock()
	invalidHandle := TheTinyChain.ParentBlock(genesisHandle)
	if !invalidHandle.IsInvalid() {
		t.Error("parent block of genesis block should be invalid handle")
	}
}

func TestHashEquality(t *testing.T) {
	hBlock0 := BlockHandle{}
	hBlock0.height = 0
	hBlock00 := BlockHandle{}
	hBlock00.height = 0
	if !hBlock0.HashSpecified() {
		t.Error("block handles must have hashes")
	}
	if hBlock0.Hash() != hBlock00.Hash() {
		t.Error("block height 0 must have same hash as block height 0")
	}
	hBlock1 := BlockHandle{}
	hBlock1.height = 1
	if hBlock0.Hash() == hBlock1.Hash() {
		t.Error("block height 1 must have different hash from block height 0")
	}
}
