package tinychain

import "testing"

func TestTransaction(t *testing.T) {
	handle := TheTinyChain.GenesisTransaction()
	trans := TheTinyChain.TransInterface(handle)
	if trans.TxiCount() != 0 {
		t.Error("genesis transaction must have no txis")
	}
	if trans.TxoCount() != 1 {
		t.Error("genesis transaction must have one txo")
	}
}

func TestThirdTransaction(t *testing.T) {
	trans := TheTinyChain.GenesisTransaction()
	nextTrans := TheTinyChain.NextTransaction(trans)
	thirdTrans := TheTinyChain.NextTransaction(nextTrans)
	transInt := TheTinyChain.TransInterface(thirdTrans)
	if transInt.TxiCount() != 1 {
		t.Error("third transaction should have one txi")
	}
	txiHandle := transInt.NthTxi(0)
	txiInt := TheTinyChain.TxiInterface(txiHandle)
	if txiInt.SourceTxo().ParentTrans() != TheTinyChain.GenesisTransaction() {
		t.Error("first txi of third transaction should be from genesis transaction")
	}
	if transInt.TxoCount() != 2 {
		t.Error("should be two txos from third transaction")
	}
	txoHandle := transInt.NthTxo(1)
	txoInt := TheTinyChain.TxoInterface(txoHandle)
	if txoInt.Satoshis() != 25 {
		t.Error("2nd txo of third transaction should be 25 satoshis")
	}
}
