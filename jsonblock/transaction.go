package jsonblock

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// jsonblock.jsonTransEssential implements chainreadinterface.ITransaction
var _ chainreadinterface.ITransaction = (*JsonTransEssential)(nil) // Check that implements

// Functions in jsonBlock.jsonTransEssential that implement chainreadinterface.ITransHandle as part of chainreadinterface.ITransaction

func (t *JsonTransEssential) Height() int64                       { return -1 }
func (t *JsonTransEssential) Hash() (indexedhashes.Sha256, error) { return t.txid, nil }
func (t *JsonTransEssential) IndicesPath() (int64, int64)         { return -1, -1 }
func (t *JsonTransEssential) HeightSpecified() bool               { return false }
func (t *JsonTransEssential) HashSpecified() bool                 { return true }
func (t *JsonTransEssential) IndicesPathSpecified() bool          { return false }
func (t *JsonTransEssential) IsTransHandle()                      {}
func (t *JsonTransEssential) IsInvalid() bool                     { return false } // A jsonTransEssential is always a valid handle

// Functions in jsonBlock.jsonTransEssential that implement chainreadinterface.ITransaction

func (t *JsonTransEssential) TxiCount() (int64, error) {
	return int64(len(t.J_vin)), nil
}
func (t *JsonTransEssential) NthTxi(n int64) (chainreadinterface.ITxiHandle, error) {
	return &t.J_vin[n], nil
}
func (t *JsonTransEssential) TxoCount() (int64, error) {
	return int64(len(t.J_vout)), nil
}
func (t *JsonTransEssential) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) {
	return &t.J_vout[n], nil
}

// jsonblock.jsonTxiEssential implements chainreadinterface.ITxi
var _ chainreadinterface.ITxi = (*JsonTxiEssential)(nil) // Check that implements

// functions in jsonblock.jsonTxiEssential to implement chainreadinterface.ITxiHandle as part of chainreadinterface.ITxi

func (txi *JsonTxiEssential) ParentTrans() chainreadinterface.ITransHandle { return &txi.parentTrans }
func (txi *JsonTxiEssential) ParentIndex() int64                           { return txi.parentVIndex }
func (txi *JsonTxiEssential) TxiHeight() int64                             { return -1 }
func (txi *JsonTxiEssential) IndicesPath() (int64, int64, int64) {
	return txi.parentTrans.blockHeight, txi.parentTrans.nthInBlock, txi.parentVIndex
}
func (txi *JsonTxiEssential) ParentSpecified() bool      { return true }
func (txi *JsonTxiEssential) TxiHeightSpecified() bool   { return false }
func (txi *JsonTxiEssential) IndicesPathSpecified() bool { return true }

// functions in jsonblock.jsonTxiEssential to implement chainreadinterface.ITxi

func (txi *JsonTxiEssential) SourceTxo() (chainreadinterface.ITxoHandle, error) {
	var txo TxoHandle
	txo.transHandle = txi.sourceTrans
	txo.vIndex = int64(txi.J_vout)
	return &txo, nil
}

// jsonblock.jsonTxoEssential implements chainreadinterface.ITxo
var _ chainreadinterface.ITxo = (*JsonTxoEssential)(nil) // Check that implements

// functions in jsonblock.jsonTxoEssential to implement chainreadinterface.ITxoHandle as part of chainreadinterface.ITxo

func (txo *JsonTxoEssential) ParentTrans() chainreadinterface.ITransHandle { return &txo.parentTrans }
func (txo *JsonTxoEssential) ParentIndex() int64                           { return txo.parentVIndex }
func (txo *JsonTxoEssential) TxoHeight() int64                             { return -1 }
func (txo *JsonTxoEssential) IndicesPath() (int64, int64, int64) {
	return txo.parentTrans.blockHeight, txo.parentTrans.nthInBlock, txo.parentVIndex
}
func (txo *JsonTxoEssential) ParentSpecified() bool      { return true }
func (txo *JsonTxoEssential) TxoHeightSpecified() bool   { return false }
func (txo *JsonTxoEssential) IndicesPathSpecified() bool { return true }

// functions in jsonblock.jsonTxoEssential to implement chainreadinterface.ITxo

func (txo *JsonTxoEssential) Satoshis() (int64, error) { return txo.satoshis, nil }
func (txo *JsonTxoEssential) Address() (chainreadinterface.IAddressHandle, error) {
	puddingHash := txo.J_scriptPubKey.puddingHash
	result := AddressHandle{}
	result.hash = puddingHash
	return &result, nil
}
func (t *JsonTransEssential) NonEssentialInts() (*map[string]int64, error) {
	return &t.nonEssentialInts, nil
}
