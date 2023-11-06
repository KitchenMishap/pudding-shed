package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"testing"
)

// Helpers that test TheTinyChain OR a copy of TheTinyChain stored in some other format.
// These helpers are passed an IBlockChain as per chainreadinterface rather than testing TheTinyChain directly.

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

func TestHashEquality_helper(hBlock0 chainreadinterface.IBlockHandle,
	hBlock00 chainreadinterface.IBlockHandle,
	hBlock1 chainreadinterface.IBlockHandle, t *testing.T) {
	if !hBlock0.HashSpecified() {
		t.Error("block handles must have hashes")
	}
	hash0, err := hBlock0.Hash()
	if err != nil {
		t.Error("cannot get hash of hBlock0")
	}
	hash00, err := hBlock00.Hash()
	if err != nil {
		t.Error("cannot get hash of hBlock00")
	}
	hash1, err := hBlock1.Hash()
	if err != nil {
		t.Error("cannot get hash of hBlock1")
	}
	if hash0 != hash00 {
		t.Error("block height 0 must have same hash as block height 0")
	}
	if hash0 == hash1 {
		t.Error("block height 1 must have different hash from block height 0")
	}
}
