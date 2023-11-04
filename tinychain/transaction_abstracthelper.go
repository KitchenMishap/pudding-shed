package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"testing"
)

func TestTransaction_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	handle := blockchain.GenesisTransaction()
	trans := blockchain.TransInterface(handle)
	if trans.TxiCount() != 0 {
		t.Error("genesis transaction must have no txis")
	}
	if trans.TxoCount() != 1 {
		t.Error("genesis transaction must have one txo")
	}
}

func TestThirdTransaction_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	trans := blockchain.GenesisTransaction()
	nextTrans := blockchain.NextTransaction(trans)
	thirdTrans := blockchain.NextTransaction(nextTrans)
	transInt := blockchain.TransInterface(thirdTrans)
	if transInt.TxiCount() != 1 {
		t.Error("third transaction should have one txi")
	}
	txiHandle := transInt.NthTxi(0)
	txiInt := blockchain.TxiInterface(txiHandle)
	if txiInt.SourceTxo().ParentTrans().Height() != blockchain.GenesisTransaction().Height() {
		t.Error("first txi of third transaction should be from genesis transaction")
	}
	if transInt.TxoCount() != 2 {
		t.Error("should be two txos from third transaction")
	}
	txoHandle := transInt.NthTxo(1)
	txoInt := blockchain.TxoInterface(txoHandle)
	if txoInt.Satoshis() != 25 {
		t.Error("2nd txo of third transaction should be 25 satoshis")
	}
}
