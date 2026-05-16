package tinychain

import (
	"testing"
)

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
	hBlock1 := BlockHandle{}
	hBlock1.height = 1
	TestHashEquality_helper(&hBlock0, &hBlock00, &hBlock1, t)
}
