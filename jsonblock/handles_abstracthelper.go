package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"testing"
)

// Helpers that test HardCodedJsonBlock[len 5] OR a copy of HardCodedJsonBlock[len 5] stored in some other format.
// These helpers are passed an IBlockChain as per chainreadinterface rather than testing HardCodedJsonBlock[] directly.

func TestGenesisHandle_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	if blockchain.IsBlockTree() {
		t.Error("blockchain can't be a BlockTree")
	}
	genesisHandle := blockchain.GenesisBlock()
	genesisBlock, err := blockchain.BlockInterface(genesisHandle)
	if err != nil {
		t.Error("could not get BlockInterface from blockchain")
	}
	if !genesisBlock.HeightSpecified() {
		t.Error("genesisBlock must specify a height")
	}
	if genesisBlock.Height() != 0 {
		t.Error("genesisBlock should be height 0")
	}
}

func TestInvalidHandle_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	genesisHandle := blockchain.GenesisBlock()
	invalidHandle := blockchain.ParentBlock(genesisHandle)
	if !invalidHandle.IsInvalid() {
		t.Error("parent block of genesis block should be invalid handle")
	}
}
