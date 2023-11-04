package tinychain

import "testing"

func TestGenesisHandle(t *testing.T) {
	TestGenesisHandle_helper(TheTinyChain, t)
}

func TestInvalidHandle(t *testing.T) {
	TestInvalidHandle_helper(TheTinyChain, t)
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
