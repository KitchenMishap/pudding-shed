package chainreadinterface

type ITransaction interface {
	TransactionHandle() HTransaction
	TxiCount() int64
	NthTxiInterface(n int64) ITxi
	TxoCount() int64
	NthTxoInterface(n int64) ITxo
}

type ITxi interface {
	SourceTransaction() HTransaction
	SourceIndex() int64
}

type ITxo interface {
	Satoshis() int64
}
