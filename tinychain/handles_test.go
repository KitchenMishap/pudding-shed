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
