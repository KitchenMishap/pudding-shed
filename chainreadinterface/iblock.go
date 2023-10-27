package chainreadinterface

type IBlock interface {
	IBlockHandle
	TransactionCount() int64
	NthTransaction(n int64) ITransHandle
}
