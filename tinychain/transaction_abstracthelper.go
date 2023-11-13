package tinychain

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"testing"
)

func TestTransaction_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	handle, err := blockchain.GenesisTransaction()
	if err != nil {
		t.Error("could not get GenesisTransaction from blockchain")
	}
	trans, err := blockchain.TransInterface(handle)
	if err != nil {
		t.Error("could not get TransInterface from blockchain")
	}
	txiCount, err := trans.TxiCount()
	if err != nil {
		t.Error("could not get TxiCount() of trans")
	}
	if txiCount != 0 {
		t.Error("genesis transaction must have no txis")
	}
	txoCount, err := trans.TxoCount()
	if err != nil {
		t.Error("could not get TxoCount() of trans")
	}
	if txoCount != 1 {
		t.Error("genesis transaction must have one txo")
	}
}

func TestThirdTransaction_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	trans, err := blockchain.GenesisTransaction()
	if err != nil {
		t.Error("could not get GenesisTransaction from blockchain")
	}
	nextTrans, err := blockchain.NextTransaction(trans)
	if err != nil {
		t.Error("could not get NextTransaction from blockchain")
	}
	thirdTrans, err := blockchain.NextTransaction(nextTrans)
	if err != nil {
		t.Error("could not get NextTransaction from blockchain")
	}
	transInt, err := blockchain.TransInterface(thirdTrans)
	if err != nil {
		t.Error("could not get TransInterface from blockchain")
	}
	txiCount, err := transInt.TxiCount()
	if err != nil {
		t.Error("could not get TxiCount of transaction")
	}
	if txiCount != 1 {
		t.Error("third transaction should have one txi")
	}
	txiHandle, err := transInt.NthTxi(0)
	if err != nil {
		t.Error("could not get NthTxi of transaction")
	}
	txiInt, err := blockchain.TxiInterface(txiHandle)
	if err != nil {
		t.Error("could not get TxiInterface from blockchain")
	}
	sourceTxo, err := txiInt.SourceTxo()
	if err != nil {
		t.Error("could not get SourceTxo() of txi")
	}
	genesisTrans, err := blockchain.GenesisTransaction()
	if err != nil {
		t.Error("could not get GenesisTransaction from blockchain")
	}
	if sourceTxo.ParentTrans().Height() != genesisTrans.Height() {
		t.Error("first txi of third transaction should be from genesis transaction")
	}
	txoCount, err := transInt.TxoCount()
	if err != nil {
		t.Error("could not get TxoCount of transaction")
	}
	if txoCount != 2 {
		t.Error("should be two txos from third transaction")
	}
	txoHandle, err := transInt.NthTxo(1)
	if err != nil {
		t.Error("could not get NthTxo of transaction")
	}
	txoInt, err := blockchain.TxoInterface(txoHandle)
	if err != nil {
		t.Error("could not get TxoInterface from blockchain")
	}
	sats, err := txoInt.Satoshis()
	if err != nil {
		t.Error("could not get satoshis from txo")
	}
	if sats != 25 {
		t.Error("2nd txo of third transaction should be 25 satoshis")
	}
}

func TestBlock2Trans2_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	block := blockchain.GenesisBlock()
	if block.IsInvalid() {
		t.Error("could not get GenesisBlock from blockchain")
	}
	nextBlock, err := blockchain.NextBlock(block)
	if err != nil {
		t.Error("could not get NextBlock from blockchain")
	}
	thirdBlock, err := blockchain.NextBlock(nextBlock)
	if err != nil {
		t.Error("could not get NextBlock (2nd) from blockchain")
	}
	blockInt, err := blockchain.BlockInterface(thirdBlock)
	if err != nil {
		t.Error("could not get BlockInterface from blockchain")
	}
	transHandle, err := blockInt.NthTransaction(1)
	if err != nil {
		t.Error("could not get transaction 1 from block 2")
	}
	transInt, err := blockchain.TransInterface(transHandle)
	if err != nil {
		t.Error("could not get transaction interface")
	}
	txiCount, err := transInt.TxiCount()
	if err != nil {
		t.Error("could not get TxiCount of transaction")
	}
	if txiCount != 1 {
		t.Error("block 2 transaction 1 should have one txi")
	}
	txiHandle, err := transInt.NthTxi(0)
	if err != nil {
		t.Error("could not get NthTxi of transaction")
	}
	txiInt, err := blockchain.TxiInterface(txiHandle)
	if err != nil {
		t.Error("could not get TxiInterface from blockchain")
	}
	sourceTxoHandle, err := txiInt.SourceTxo()
	if err != nil {
		t.Error("could not get sourceTxo handle from txi interface")
	}
	sourceTxoTrans := sourceTxoHandle.ParentTrans()
	sourceTransHeight := sourceTxoTrans.Height()
	if sourceTransHeight != 5 {
		t.Error("source trans should be height 5")
	}
}
