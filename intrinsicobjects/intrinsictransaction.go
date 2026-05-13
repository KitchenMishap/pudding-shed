package intrinsicobjects

import (
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// Intrinsic objects may not hold data inferred from external representations.
// For example, a transaction may refer to the txid of a previous transaction, because that txid is available in the transaction binary.
// But a transaction may not hold an index addressing a previous transaction, because that index is a consequence of the order of transactions.

type Transaction struct {
	TxId            indexedhashes.Sha256
	Version         uint32
	IsSegWit        bool
	BitcoinCoreTxis []Txi // "Bitcoin Core" to remind us these are DIFFERENT to "Pudding Shed's" Txi's
	Txos            []Txo // (Bitcoin Core has Txi's for coinbase transactions; Pudding Shed DISCARDS them)

	Size         int
	Weight       int
	VSize        int
	StrippedSize int
}

type Txi struct {
	TxId indexedhashes.Sha256
	VOut int64
}

type Txo struct {
	Value        int64
	ScriptPubKey []byte
}
