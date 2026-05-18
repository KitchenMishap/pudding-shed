package intrinsicobjects

import (
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Intrinsic objects may not hold data inferred from external representations.
// For example, a transaction may refer to the txid of a previous transaction, because that txid is available in the transaction binary.
// But a transaction may not hold an index addressing a previous transaction, because that index is a consequence of the order of transactions.

type Transaction struct {
	TxId     indexedhashes.Sha256
	Version  uint32
	IsSegWit bool
	TxisTEMP []Txi
	TxosTEMP []Txo

	Size         int
	Weight       int
	VSize        int
	StrippedSize int
}

func (t *Transaction) TxiCount() int64 { return int64(len(t.TxisTEMP)) }
func (t *Transaction) TxoCount() int64 { return int64(len(t.TxosTEMP)) }
func (t *Transaction) GetTxi(storage *MultiTransactionStorage, index int64) Txi {
	return t.TxisTEMP[index]
}
func (t *Transaction) GetTxo(storage *MultiTransactionStorage, index int64) Txo {
	return t.TxosTEMP[index]
}

type Txi struct {
	TxId indexedhashes.Sha256
	VOut int64
}

type Txo struct {
	Value        int64
	ScriptPubKey []byte
}
