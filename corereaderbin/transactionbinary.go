package corereaderbin

import (
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

type TransactionBinary struct {
	Txid     indexedhashes.Sha256
	Version  uint32
	IsSegWit bool
	Txis     []TxiBinary
	Txos     []TxoBinary
}

type TxiBinary struct {
	TxId indexedhashes.Sha256
	Vout uint32
}

type TxoBinary struct {
	Value        uint64
	PuddingHash3 indexedhashes.Sha256
}
