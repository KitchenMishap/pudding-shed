package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"testing"
)

func TestJustFiveCoinbaseBlocks_helper(blockchain chainreadinterface.IBlockChain, t *testing.T) {
	// Block handle
	blockHandle := blockchain.GenesisBlock()
	for i := 0; i < 5; i++ {
		if blockHandle.IsInvalid() {
			t.Error("block handle cannot be invalid")
		}

		// Block interface
		block, err := blockchain.BlockInterface(blockHandle)
		if err != nil {
			t.Error("could not get BlockInterface from blockchain")
		}
		if block == nil {
			t.Error("block cannot be nil")
		}
		if block.IsInvalid() {
			t.Error("block cannot be invalid")
		}

		// One transaction
		count, err := block.TransactionCount()
		if count != 1 {
			t.Error("block should have exactly one transaction")
		}
		transHandle, err := block.NthTransaction(0)
		if err != nil {
			t.Error(err)
		}
		trans, err := blockchain.TransInterface(transHandle)
		if err != nil {
			t.Error(err)
		}
		if trans == nil {
			t.Error("transaction cannot be nil")
		}

		// 0 Txis
		count, err = trans.TxiCount()
		if count != 0 {
			t.Error("coinbase transaction should have zero inputs")
		}

		// 1 Txo
		count, err = trans.TxoCount()
		if count != 1 {
			t.Error("transaction should have one output")
		}
		txohandle, err := trans.NthTxo(0)
		if err != nil {
			t.Error(err)
		}
		if txohandle == nil {
			t.Error("txohandle cannot be nil")
		}
		txo, err := blockchain.TxoInterface(txohandle)
		if err != nil {
			t.Error(err)
		}
		if txo == nil {
			t.Error("txo cannot be nil")
		}
		sats, err := txo.Satoshis()
		if err != nil {
			t.Error(err)
		}
		if sats != int64(50*100_000_000) {
			t.Error("satoshis should be 50 bitcoins")
		}

		// next block
		blockHandle, err = blockchain.NextBlock(blockHandle)
		if err != nil {
			t.Error(err)
		}
	}
	if !blockHandle.IsInvalid() {
		t.Error("next block handle after last should be invalid")
	}
}
