package tinychain

import "testing"

func TestTransaction(t *testing.T) {
	transHandle := TheTinyChain.GenesisTransaction()
	transInt := TheTinyChain.TransactionInterface(transHandle)
	if transInt.TxiCount() != 0 {
		t.Error()
	}
	if transInt.TxoCount() != 1 {
		t.Error()
	}
}

func TestThirdTransaction(t *testing.T) {
	transHandle := TheTinyChain.GenesisTransaction()
	nextTrans := TheTinyChain.NextTransaction(transHandle)
	thirdTrans := TheTinyChain.NextTransaction(nextTrans)
	transInt := TheTinyChain.TransactionInterface(thirdTrans)
	if transInt.TxiCount() != 1 {
		t.Error()
	}
	atxi := transInt.NthTxiInterface(0)
	if atxi.SourceTransaction() != TheTinyChain.GenesisTransaction() {
		t.Error()
	}
	if transInt.TxoCount() != 2 {
		t.Error()
	}
	atxo := transInt.NthTxoInterface(1)
	if atxo.Satoshis() != 25 {
		t.Error()
	}
}
