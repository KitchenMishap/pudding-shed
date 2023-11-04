package tinychain

import "testing"

func TestTransaction(t *testing.T) {
	TestTransaction_helper(TheTinyChain, t)
}

func TestThirdTransaction(t *testing.T) {
	TestThirdTransaction_helper(TheTinyChain, t)
}
