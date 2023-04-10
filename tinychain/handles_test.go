package tinychain

import "testing"

func TestGenesisHandle(t *testing.T) {
	genesisHandle := TheTinyChain.GenesisBlock()
	genesisHeight := TheHandles.HeightFromHBlock(genesisHandle)
	if genesisHeight != 0 {
		t.Error()
	}
	if TheHandles.HBlockFromHeight(genesisHeight) != genesisHandle {
		t.Error()
	}
}

func TestInvalidHandle(t *testing.T) {
	genesisHandle := TheTinyChain.GenesisBlock()
	invalidHandle := TheTinyChain.ParentBlock(genesisHandle)
	if !invalidHandle.IsInvalid() {
		t.Error()
	}
}
