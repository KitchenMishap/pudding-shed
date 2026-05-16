package tinychain

import "testing"

func TestTransaction(t *testing.T) {
	TestTransaction_helper(TheTinyChain, t)
}

func TestThirdTransaction(t *testing.T) {
	TestThirdTransaction_helper(TheTinyChain, t)
}

func TestBlock2Trans2(t *testing.T) {
	TestBlock2Trans2_helper(TheTinyChain, t)
}
