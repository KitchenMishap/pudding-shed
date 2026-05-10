package intrinsicobjectscri

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"github.com/KitchenMishap/pudding-shed/intrinsicobjects"
)

// A Transaction object with an intrinsicobjects.Transaction, with adornments to
// implement chainreadinterface Interfaces

type Transaction struct {
	intrinsic *intrinsicobjects.Transaction
	txis      []Txi
	txos      []Txo
	neisMap   map[string]int64 // Non Essential Ints
}

func NewTransaction(intrinsic *intrinsicobjects.Transaction) *Transaction {
	result := Transaction{}
	result.intrinsic = intrinsic

	result.txis = make([]Txi, len(intrinsic.Txis))
	for i := range intrinsic.Txis {
		result.txis[i].intrinsic = &intrinsic.Txis[i]
		result.txis[i].parentTransaction = &result
		result.txis[i].parentIndex = int64(i)
	}

	result.txos = make([]Txo, len(intrinsic.Txos))
	for i := range intrinsic.Txos {
		result.txos[i].intrinsic = &intrinsic.Txos[i]
		result.txos[i].parentTransaction = &result
		result.txos[i].parentIndex = int64(i)
	}

	result.neisMap = make(map[string]int64)
	// ToDo

	return &result
}

type Txi struct {
	intrinsic         *intrinsicobjects.Txi
	parentTransaction *Transaction
	parentIndex       int64
}

type Txo struct {
	intrinsic         *intrinsicobjects.Txo
	parentTransaction *Transaction
	parentIndex       int64
}

// intrinsicobjectscri.Transaction implements chainreadinterface.ITransaction
var _ chainreadinterface.ITransaction = (*Transaction)(nil) // Check that implements

func (t *Transaction) TxiCount() (int64, error)                              { return int64(len(t.txis)), nil }
func (t *Transaction) NthTxi(n int64) (chainreadinterface.ITxiHandle, error) { return &t.txis[n], nil }
func (t *Transaction) TxoCount() (int64, error)                              { return int64(len(t.txos)), nil }
func (t *Transaction) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) { return &t.txos[n], nil }
func (t *Transaction) NonEssentialInts() (*map[string]int64, error)          { return &t.neisMap, nil }
func (t *Transaction) AllTxoSatoshis() ([]int64, error) {
	result := make([]int64, len(t.txos))
	for i := range len(t.txos) {
		result[i] = t.txos[i].intrinsic.Value
	}
	return result, nil
}

// intrinsicobjectscri.Transaction also implements chainreadinterface.ITransHandle
var _ chainreadinterface.ITransHandle = (*Transaction)(nil) // Check that implements

func (t *Transaction) Height() int64                       { return -1 }
func (t *Transaction) HeightSpecified() bool               { return false }
func (t *Transaction) Hash() (indexedhashes.Sha256, error) { return t.intrinsic.TxId, nil }
func (t *Transaction) HashSpecified() bool                 { return true }
func (t *Transaction) IsTransHandle()                      {}
func (t *Transaction) IsInvalid() bool                     { return false }
func (t *Transaction) IndicesPath() (int64, int64)         { return -1, -1 }
func (t *Transaction) IndicesPathSpecified() bool          { return false }

// intrinsicobjectscri.Txi implements chainreadinterface.ITxi
var _ chainreadinterface.ITxi = (*Txi)(nil) // Check that implements

func (txi *Txi) SourceTxo() (chainreadinterface.ITxoHandle, error) {
	hTxo := TxoHandle{}
	hTxo.txId = txi.intrinsic.TxId
	hTxo.vIndex = txi.intrinsic.VOut
	return &hTxo, nil
}

// intrinsicobjectscri.Txi also implements chainreadinterface.ITxiHandle
var _ chainreadinterface.ITxiHandle = (*Txi)(nil) // Check that implements

func (txi *Txi) TxiHeight() int64                   { return -1 }
func (txi *Txi) TxiHeightSpecified() bool           { return false }
func (txi *Txi) IsInvalid() bool                    { return false }
func (txi *Txi) IndicesPath() (int64, int64, int64) { return -1, -1, -1 }
func (txi *Txi) IndicesPathSpecified() bool         { return false }
func (txi *Txi) ParentTrans() chainreadinterface.ITransHandle {
	result := TransHandle{}
	result.transactionId = txi.parentTransaction.intrinsic.TxId
	result.isInvalid = false
	return &result
}
func (txi *Txi) ParentIndex() int64    { return txi.parentIndex }
func (txi *Txi) ParentSpecified() bool { return true }

// intrinsicobjectscri.Txo implements chainreadinterface.ITxo
var _ chainreadinterface.ITxo = (*Txo)(nil) // Check that implements

func (txo *Txo) Satoshis() (int64, error) { return txo.intrinsic.Value, nil }
func (txo *Txo) Address() (chainreadinterface.IAddressHandle, error) {
	result := AddressHandle{}
	result.puddingHash3 = txo.intrinsic.AddressPuddingHash3
	return &result, nil
}

// intrinsicobjects.Txo also implements chainreadinterface.ITxoHandle
var _ chainreadinterface.ITxoHandle = (*Txo)(nil) // Check that implements

func (txo *Txo) TxoHeight() int64                   { return -1 }
func (txo *Txo) TxoHeightSpecified() bool           { return false }
func (txo *Txo) IsInvalid() bool                    { return false }
func (txo *Txo) IndicesPath() (int64, int64, int64) { return -1, -1, -1 }
func (txo *Txo) IndicesPathSpecified() bool         { return false }
func (txo *Txo) ParentTrans() chainreadinterface.ITransHandle {
	result := TransHandle{}
	result.transactionId = txo.parentTransaction.intrinsic.TxId
	result.isInvalid = false
	return &result
}
func (txo *Txo) ParentIndex() int64    { return txo.parentIndex }
func (txo *Txo) ParentSpecified() bool { return true }
