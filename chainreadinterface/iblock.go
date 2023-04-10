package chainreadinterface

type IBlock interface {
	BlockHandle() HBlock
	BlockHeight() int64
	TransactionCount() int64
	NthTransactionHandle(n int64) HTransaction
}
