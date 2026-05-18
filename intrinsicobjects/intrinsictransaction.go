package intrinsicobjects

import (
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Intrinsic objects may not hold data inferred from external representations.
// For example, a transaction may refer to the txid of a previous transaction, because that txid is available in the transaction binary.
// But a transaction may not hold an index addressing a previous transaction, because that index is a consequence of the order of transactions.

type Transaction struct {
	TxId      indexedhashes.Sha256
	Version   uint32
	IsSegWit  bool
	TxisStart int64 // An index into the supplied MultiTransactionStorage
	TxisCount int64
	TxosStart int64 // An index into the supplied MultiTransactionStorage
	TxosCount int64

	Size         int
	Weight       int
	VSize        int
	StrippedSize int
}

func (t *Transaction) TxiCount() int64 { return t.TxisCount }
func (t *Transaction) TxoCount() int64 { return t.TxosCount }
func (t *Transaction) GetTxi(storage *MultiTransactionStorage, index int64) Txi {
	return storage.Txis[t.TxisStart+index]
}
func (t *Transaction) GetTxo(storage *MultiTransactionStorage, index int64) Txo {
	return storage.Txos[t.TxosStart+index]
}

type Txi struct {
	TxId indexedhashes.Sha256
	VOut int64
}

type Txo struct {
	Value        int64
	ScriptPubKey []byte
}
