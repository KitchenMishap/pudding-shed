package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// jsonblock.jsonTransEssential implements chainreadinterface.ITransaction
var _ chainreadinterface.ITransaction = (*jsonTransEssential)(nil) // Check that implements

// Functions in jsonBlock.jsonTransEssential that implement chainreadinterface.ITransHandle as part of chainreadinterface.ITransaction

func (t *jsonTransEssential) Height() int64                       { return -1 }
func (t *jsonTransEssential) Hash() (indexedhashes.Sha256, error) { return t.txid, nil }
func (t *jsonTransEssential) IndicesPath() (int64, int64)         { return -1, -1 }
func (t *jsonTransEssential) HeightSpecified() bool               { return false }
func (t *jsonTransEssential) HashSpecified() bool                 { return true }
func (t *jsonTransEssential) IndicesPathSpecified() bool          { return false }
func (t *jsonTransEssential) IsTransHandle()                      {}
func (t *jsonTransEssential) IsInvalid() bool                     { return false } // A jsonTransEssential is always a valid handle

// Functions in jsonBlock.jsonTransEssential that implement chainreadinterface.ITransaction

func (t *jsonTransEssential) TxiCount() (int64, error) {
	return int64(len(t.J_vin)), nil
}
func (t *jsonTransEssential) NthTxi(n int64) (chainreadinterface.ITxiHandle, error) {
	return &t.J_vin[n], nil
}
func (t *jsonTransEssential) TxoCount() (int64, error) {
	return int64(len(t.J_vout)), nil
}
func (t *jsonTransEssential) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) {
	return &t.J_vout[n], nil
}

// jsonblock.jsonTxiEssential implements chainreadinterface.ITxi
var _ chainreadinterface.ITxi = (*jsonTxiEssential)(nil) // Check that implements

// functions in jsonblock.jsonTxiEssential to implement chainreadinterface.ITxiHandle as part of chainreadinterface.ITxi

func (txi *jsonTxiEssential) ParentTrans() chainreadinterface.ITransHandle { return &txi.parentTrans }
func (txi *jsonTxiEssential) ParentIndex() int64                           { return txi.parentVIndex }
func (txi *jsonTxiEssential) TxiHeight() int64                             { return -1 }
func (txi *jsonTxiEssential) IndicesPath() (int64, int64, int64) {
	return txi.parentTrans.blockHeight, txi.parentTrans.nthInBlock, txi.parentVIndex
}
func (txi *jsonTxiEssential) ParentSpecified() bool      { return true }
func (txi *jsonTxiEssential) TxiHeightSpecified() bool   { return false }
func (txi *jsonTxiEssential) IndicesPathSpecified() bool { return true }

// functions in jsonblock.jsonTxiEssential to implement chainreadinterface.ITxi

func (txi *jsonTxiEssential) SourceTxo() (chainreadinterface.ITxoHandle, error) {
	var txo TxoHandle
	txo.transHandle = txi.sourceTrans
	txo.vIndex = int64(txi.J_vout)
	return &txo, nil
}

// jsonblock.jsonTxoEssential implements chainreadinterface.ITxo
var _ chainreadinterface.ITxo = (*jsonTxoEssential)(nil) // Check that implements

// functions in jsonblock.jsonTxoEssential to implement chainreadinterface.ITxoHandle as part of chainreadinterface.ITxo

func (txo *jsonTxoEssential) ParentTrans() chainreadinterface.ITransHandle { return &txo.parentTrans }
func (txo *jsonTxoEssential) ParentIndex() int64                           { return txo.parentVIndex }
func (txo *jsonTxoEssential) TxoHeight() int64                             { return -1 }
func (txo *jsonTxoEssential) IndicesPath() (int64, int64, int64) {
	return txo.parentTrans.blockHeight, txo.parentTrans.nthInBlock, txo.parentVIndex
}
func (txo *jsonTxoEssential) ParentSpecified() bool      { return true }
func (txo *jsonTxoEssential) TxoHeightSpecified() bool   { return false }
func (txo *jsonTxoEssential) IndicesPathSpecified() bool { return true }

// functions in jsonblock.jsonTxoEssential to implement chainreadinterface.ITxo

func (txo *jsonTxoEssential) Satoshis() (int64, error) { return txo.satoshis, nil }
