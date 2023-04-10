package chainreadinterface

type IBlock interface {
	BlockHandle() HBlock
	BlockHeight() int64
	BlockHash() [32]byte
	TransactionCount() int64
	NthTransactionHandle(n int64) HTransaction
}
