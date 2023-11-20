package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// jsonblock.Transaction implements chainreadinterface.ITransaction
var _ chainreadinterface.ITransaction = (*Transaction)(nil) // Check that implements
type Transaction struct {
	json jsonTransEssential
}

// Functions in jsonBlock.Transaction that implement chainreadinterface.ITransHandle as part of chainreadinterface.ITransaction

func (t *Transaction) Height() int64                       { return -1 }
func (t *Transaction) Hash() (indexedhashes.Sha256, error) { return t.json.Txid, nil }
func (t *Transaction) HeightSpecified() bool               { return false }
func (t *Transaction) HashSpecified() bool                 { return true }
func (t *Transaction) IsTransHandle()                      {}
func (t *Transaction) IsInvalid() bool                     { return false } // A Transaction is always a valid handle

// Functions in jsonBlock.Transaction that implement chainreadinterface.ITransaction

func (t *Transaction) TxiCount() (int64, error) { return int64(len(t.json.Vin)), nil }
func (t *Transaction) NthTxi(n int64) (chainreadinterface.ITxiHandle, error) {
	var txi Txi
	txi.parentTrans = t
	txi.parentIndex = n
	return &txi, nil
}
func (t *Transaction) TxoCount() (int64, error) { return int64(len(t.json.Vout)), nil }
func (t *Transaction) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) {
	var txo TxoHandle
	txo.txid = t.json.Txid
	txo.vout = n
	return &txo, nil
}

// jsonblock.Txi implements chainreadinterface.ITxi and is also used for chainreadinterface.ITxiHandle
var _ chainreadinterface.ITxi = (*Txi)(nil) // Check that implements
type Txi struct {
	parentTrans *Transaction // A jsonblock.Txi only exists in the context of a transaction, so just use ptr
	parentIndex int64
}

// functions in jsonblock.Txi to implement chainreadinterface.ITxiHandle

func (txi *Txi) ParentTrans() chainreadinterface.ITransHandle { return txi.parentTrans }
func (txi *Txi) ParentIndex() int64                           { return txi.parentIndex }
func (txi *Txi) TxiHeight() int64                             { return -1 }
func (txi *Txi) ParentSpecified() bool                        { return true }
func (txi *Txi) TxiHeightSpecified() bool                     { return false }

// functions in jsonblock.Txi to implement chainreadinterface.ITxi

func (txi *Txi) SourceTxo() (chainreadinterface.ITxoHandle, error) {
	var txo TxoHandle
	vinElement := &txi.parentTrans.json.Vin[txi.parentIndex]
	txo.txid = vinElement.Txid
	txo.vout = vinElement.Vout
	return &txo, nil
}

// jsonblock.Txo implements chainreadinterface.ITxo
var _ chainreadinterface.ITxo = (*Txo)(nil) // Check that implements
type Txo struct {
	parentTrans *Transaction // A jsonblock.Txo only exists in the context of a transaction, so just use ptr
	parentIndex int64
}

// functions in jsonblock.Txo to implement chainreadinterface.ITxoHandle as part of chainreadinterface.ITxo

func (txo *Txo) ParentTrans() chainreadinterface.ITransHandle { return txo.parentTrans }
func (txo *Txo) ParentIndex() int64                           { return txo.parentIndex }
func (txo *Txo) TxoHeight() int64                             { return -1 }
func (txo *Txo) ParentSpecified() bool                        { return true }
func (txo *Txo) TxoHeightSpecified() bool                     { return false }

// functions in jsonblock.Txo to implement chainreadinterface.ITxo

func (txo *Txo) Satoshis() (int64, error) {
	const SatoshisPerBitcoin = float64(100_000_000)
	return int64(txo.parentTrans.json.Vout[txo.parentIndex].Value * SatoshisPerBitcoin), nil
}
