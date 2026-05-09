package intrinsicobjects

import (
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Intrinsic objects may not hold data inferred from external representations.
// For example, a transaction may refer to the txid of a previous transaction, because that txid is available in the transaction binary.
// But a transaction may not hold an index addressing a previous transaction, because that index is a consequence of the order of transactions.

type Transaction struct {
	TxId     indexedhashes.Sha256
	Version  uint32
	IsSegWit bool
	Txis     []Txi
	Txos     []Txo

	NonEssentialIntsMap map[string]int64 // ToDo Populate this
}

// intrinsicobjects.Transaction implements chainreadinterface.ITransaction
var _ chainreadinterface.ITransaction = (*Transaction)(nil) // Check that implements

func (t *Transaction) TxiCount() (int64, error)                              { return int64(len(t.Txis)), nil }
func (t *Transaction) NthTxi(n int64) (chainreadinterface.ITxiHandle, error) { return &t.Txis[n], nil }
func (t *Transaction) TxoCount() (int64, error)                              { return int64(len(t.Txos)), nil }
func (t *Transaction) NthTxo(n int64) (chainreadinterface.ITxoHandle, error) { return &t.Txos[n], nil }
func (t *Transaction) NonEssentialInts() (*map[string]int64, error) {
	return &t.NonEssentialIntsMap, nil
}
func (t *Transaction) AllTxoSatoshis() ([]int64, error) {
	result := make([]int64, len(t.Txos))
	for i := range len(t.Txos) {
		result[i] = t.Txos[i].Value
	}
	return result, nil
}

// intrincicobjects.Transaction also implements chainreadinterface.ITransHandle
var _ chainreadinterface.ITransHandle = (*Transaction)(nil) // Check that implements

func (t *Transaction) Height() int64                       { return -1 }
func (t *Transaction) HeightSpecified() bool               { return false }
func (t *Transaction) Hash() (indexedhashes.Sha256, error) { return t.TxId, nil }
func (t *Transaction) HashSpecified() bool                 { return true }
func (t *Transaction) IsTransHandle()                      {}
func (t *Transaction) IsInvalid() bool                     { return false }
func (t *Transaction) IndicesPath() (int64, int64)         { return -1, -1 }
func (t *Transaction) IndicesPathSpecified() bool          { return false }

type Txi struct {
	TxId indexedhashes.Sha256
	VOut int64
}

// intrinsicobjects.Txi implements chainreadinterface.ITxi
var _ chainreadinterface.ITxi = (*Txi)(nil) // Check that implements

func (txi *Txi) SourceTxo() (chainreadinterface.ITxoHandle, error) {
	hTxo := TxoHandle{}
	hTxo.txId = txi.TxId
	hTxo.vIndex = txi.VOut
	return &hTxo, nil
}

// intrinsicobjects.Txi also implements chainreadinterface.ITxiHandle
var _ chainreadinterface.ITxiHandle = (*Txi)(nil) // Check that implements

func (txi *Txi) TxiHeight() int64                             { return -1 }
func (txi *Txi) TxiHeightSpecified() bool                     { return false }
func (txi *Txi) IsInvalid() bool                              { return false }
func (txi *Txi) IndicesPath() (int64, int64, int64)           { return -1, -1, -1 }
func (txi *Txi) IndicesPathSpecified() bool                   { return false }
func (txi *Txi) ParentTrans() chainreadinterface.ITransHandle { return nil }
func (txi *Txi) ParentIndex() int64                           { return -1 }
func (txi *Txi) ParentSpecified() bool                        { return false }

type Txo struct {
	Value               int64
	AddressPuddingHash3 indexedhashes.Sha256 // PuddingHash3 is not a hash familiar to bitcoiners (peculiar to pudding-shed software)
}

// intrinsicobjects.Txo implements chainreadinterface.ITxo
var _ chainreadinterface.ITxo = (*Txo)(nil) // Check that implements
func (txo *Txo) Satoshis() (int64, error)   { return txo.Value, nil }
func (txo *Txo) Address() (chainreadinterface.IAddressHandle, error) {
	result := AddressHandle{}
	result.puddingHash3 = txo.AddressPuddingHash3
	return &result, nil
}

// intrinsicobjects.Txo also implements chainreadinterface.ITxoHandle
var _ chainreadinterface.ITxoHandle = (*Txo)(nil) // Check that implements

func (txo *Txo) TxoHeight() int64                             { return -1 }
func (txo *Txo) TxoHeightSpecified() bool                     { return false }
func (txo *Txo) IsInvalid() bool                              { return false }
func (txo *Txo) IndicesPath() (int64, int64, int64)           { return -1, -1, -1 }
func (txo *Txo) IndicesPathSpecified() bool                   { return false }
func (txo *Txo) ParentTrans() chainreadinterface.ITransHandle { return nil }
func (txo *Txo) ParentIndex() int64                           { return -1 }
func (txo *Txo) ParentSpecified() bool                        { return false }
