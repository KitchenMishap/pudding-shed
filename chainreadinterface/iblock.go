package chainreadinterface

type IBlock interface {
	IBlockHandle
	TransactionCount() (int64, error)
	NthTransaction(n int64) (ITransHandle, error)
}
